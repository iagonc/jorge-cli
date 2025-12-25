package health

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// HTTPChecker interface for testability
type HTTPChecker interface {
	Check(ctx context.Context, target models.HealthTarget) models.HealthCheckResult
}

// DefaultHTTPChecker implements HTTPChecker
type DefaultHTTPChecker struct {
	Client *http.Client
}

func (c *DefaultHTTPChecker) Check(ctx context.Context, target models.HealthTarget) models.HealthCheckResult {
	result := models.HealthCheckResult{
		Name:      target.Name,
		URL:       target.URL,
		CheckedAt: time.Now(),
	}

	timeout := time.Duration(target.Timeout) * time.Second
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	method := target.Method
	if method == "" {
		method = http.MethodGet
	}

	var bodyReader io.Reader
	if target.Body != "" {
		bodyReader = strings.NewReader(target.Body)
	}

	req, err := http.NewRequestWithContext(ctx, method, target.URL, bodyReader)
	if err != nil {
		result.Status = models.StatusError
		result.Error = fmt.Sprintf("Failed to create request: %v", err)
		return result
	}

	for key, value := range target.Headers {
		req.Header.Set(key, value)
	}

	start := time.Now()
	resp, err := c.Client.Do(req)
	result.ResponseTime = time.Since(start)

	if err != nil {
		result.Status = models.StatusError
		result.Error = fmt.Sprintf("Request failed: %v", err)
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode

	// Check expected status
	expectedStatus := target.ExpectedStatus
	if expectedStatus == 0 {
		expectedStatus = http.StatusOK
	}

	// Check body if expected
	if target.ExpectedBody != "" {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			result.Status = models.StatusError
			result.Error = fmt.Sprintf("Failed to read body: %v", err)
			return result
		}
		result.BodyMatch = strings.Contains(string(body), target.ExpectedBody)
	} else {
		result.BodyMatch = true
	}

	// Determine health status
	if resp.StatusCode == expectedStatus && result.BodyMatch {
		result.Status = models.StatusHealthy
	} else {
		result.Status = models.StatusUnhealthy
		if resp.StatusCode != expectedStatus {
			result.Error = fmt.Sprintf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
		} else if !result.BodyMatch {
			result.Error = "Body does not match expected content"
		}
	}

	return result
}

// HealthCheckUsecase handles health check operations
type HealthCheckUsecase struct {
	Logger  *zap.Logger
	Checker HTTPChecker
}

// NewHealthCheckUsecase creates a new HealthCheckUsecase
func NewHealthCheckUsecase(logger *zap.Logger) *HealthCheckUsecase {
	return &HealthCheckUsecase{
		Logger: logger,
		Checker: &DefaultHTTPChecker{
			Client: &http.Client{
				Timeout: 30 * time.Second,
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					if len(via) >= 10 {
						return fmt.Errorf("too many redirects")
					}
					return nil
				},
			},
		},
	}
}

// LoadConfig loads health check configuration from a YAML file
func (u *HealthCheckUsecase) LoadConfig(configPath string) (*models.HealthCheckConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config models.HealthCheckConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// RunHealthChecks runs all health checks from the config
func (u *HealthCheckUsecase) RunHealthChecks(ctx context.Context, config *models.HealthCheckConfig, parallel bool) (*models.HealthCheckSummary, []error) {
	u.Logger.Info("Starting health checks", zap.Int("count", len(config.Checks)))

	startTime := time.Now()
	summary := &models.HealthCheckSummary{
		TotalChecks: len(config.Checks),
	}

	var errorsList []error

	if parallel {
		summary.Results, errorsList = u.runParallel(ctx, config.Checks)
	} else {
		summary.Results, errorsList = u.runSequential(ctx, config.Checks)
	}

	// Calculate summary
	for _, result := range summary.Results {
		switch result.Status {
		case models.StatusHealthy:
			summary.Healthy++
		case models.StatusUnhealthy:
			summary.Unhealthy++
		case models.StatusError:
			summary.Errors++
		}
	}

	summary.Duration = time.Since(startTime)

	u.Logger.Info("Health checks completed",
		zap.Int("healthy", summary.Healthy),
		zap.Int("unhealthy", summary.Unhealthy),
		zap.Int("errors", summary.Errors),
		zap.Duration("duration", summary.Duration))

	return summary, errorsList
}

func (u *HealthCheckUsecase) runParallel(ctx context.Context, targets []models.HealthTarget) ([]models.HealthCheckResult, []error) {
	var (
		wg         sync.WaitGroup
		mu         sync.Mutex
		results    []models.HealthCheckResult
		errorsList []error
	)

	for _, target := range targets {
		wg.Add(1)
		go func(t models.HealthTarget) {
			defer wg.Done()

			result := u.Checker.Check(ctx, t)

			mu.Lock()
			results = append(results, result)
			if result.Error != "" {
				errorsList = append(errorsList, fmt.Errorf("%s: %s", t.Name, result.Error))
			}
			mu.Unlock()
		}(target)
	}

	wg.Wait()
	return results, errorsList
}

func (u *HealthCheckUsecase) runSequential(ctx context.Context, targets []models.HealthTarget) ([]models.HealthCheckResult, []error) {
	var results []models.HealthCheckResult
	var errorsList []error

	for _, target := range targets {
		select {
		case <-ctx.Done():
			return results, append(errorsList, ctx.Err())
		default:
		}

		result := u.Checker.Check(ctx, target)
		results = append(results, result)
		if result.Error != "" {
			errorsList = append(errorsList, fmt.Errorf("%s: %s", target.Name, result.Error))
		}
	}

	return results, errorsList
}

// CheckSingle checks a single URL
func (u *HealthCheckUsecase) CheckSingle(ctx context.Context, url string, expectedStatus int) (*models.HealthCheckResult, error) {
	target := models.HealthTarget{
		Name:           url,
		URL:            url,
		Method:         http.MethodGet,
		ExpectedStatus: expectedStatus,
		Timeout:        10,
	}

	result := u.Checker.Check(ctx, target)
	return &result, nil
}
