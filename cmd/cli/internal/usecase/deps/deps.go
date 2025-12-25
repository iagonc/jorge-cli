package deps

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// DepsUsecase handles dependency checking
type DepsUsecase struct {
	logger *zap.Logger
}

// NewDepsUsecase creates a new DepsUsecase
func NewDepsUsecase(logger *zap.Logger) *DepsUsecase {
	return &DepsUsecase{logger: logger}
}

// LoadConfig loads dependencies config from a file
func (u *DepsUsecase) LoadConfig(path string) (*models.DepsConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config models.DepsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// CheckAll checks all dependencies
func (u *DepsUsecase) CheckAll(ctx context.Context, config *models.DepsConfig) (*models.DepsResult, error) {
	startTime := time.Now()

	result := &models.DepsResult{
		ConfigName: config.Name,
		TotalDeps:  len(config.Dependencies),
		Results:    make([]models.DependencyResult, len(config.Dependencies)),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, dep := range config.Dependencies {
		wg.Add(1)
		go func(idx int, dep models.Dependency) {
			defer wg.Done()
			depResult := u.checkDependency(ctx, dep)
			mu.Lock()
			result.Results[idx] = depResult
			if depResult.Healthy {
				result.HealthyDeps++
			} else {
				result.UnhealthyDeps++
			}
			mu.Unlock()
		}(i, dep)
	}

	wg.Wait()

	result.Duration = time.Since(startTime)
	result.AllHealthy = result.UnhealthyDeps == 0

	return result, nil
}

// CheckSingle checks a single dependency
func (u *DepsUsecase) CheckSingle(ctx context.Context, dep models.Dependency) models.DependencyResult {
	return u.checkDependency(ctx, dep)
}

func (u *DepsUsecase) checkDependency(ctx context.Context, dep models.Dependency) models.DependencyResult {
	result := models.DependencyResult{
		Name:     dep.Name,
		Type:     dep.Type,
		Required: dep.Required,
	}

	timeout := time.Duration(dep.Timeout) * time.Second
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()

	switch dep.Type {
	case models.DepTypeHTTP, "":
		result = u.checkHTTP(ctx, dep, result)
	case models.DepTypeTCP:
		result = u.checkTCP(ctx, dep, result)
	case models.DepTypePostgres:
		result = u.checkTCP(ctx, models.Dependency{
			Name: dep.Name,
			Host: dep.Host,
			Port: func() int { if dep.Port == 0 { return 5432 }; return dep.Port }(),
		}, result)
		result.Type = models.DepTypePostgres
	case models.DepTypeMySQL:
		result = u.checkTCP(ctx, models.Dependency{
			Name: dep.Name,
			Host: dep.Host,
			Port: func() int { if dep.Port == 0 { return 3306 }; return dep.Port }(),
		}, result)
		result.Type = models.DepTypeMySQL
	case models.DepTypeRedis:
		result = u.checkTCP(ctx, models.Dependency{
			Name: dep.Name,
			Host: dep.Host,
			Port: func() int { if dep.Port == 0 { return 6379 }; return dep.Port }(),
		}, result)
		result.Type = models.DepTypeRedis
	case models.DepTypeMongoDB:
		result = u.checkTCP(ctx, models.Dependency{
			Name: dep.Name,
			Host: dep.Host,
			Port: func() int { if dep.Port == 0 { return 27017 }; return dep.Port }(),
		}, result)
		result.Type = models.DepTypeMongoDB
	case models.DepTypeDNS:
		result = u.checkDNS(ctx, dep, result)
	default:
		result.Error = fmt.Sprintf("unknown dependency type: %s", dep.Type)
	}

	result.ResponseTime = time.Since(start)
	return result
}

func (u *DepsUsecase) checkHTTP(ctx context.Context, dep models.Dependency, result models.DependencyResult) models.DependencyResult {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	url := dep.URL
	if url == "" && dep.Host != "" {
		port := dep.Port
		if port == 0 {
			port = 80
		}
		url = fmt.Sprintf("http://%s:%d", dep.Host, port)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	for k, v := range dep.Headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer resp.Body.Close()

	expectedStatus := dep.ExpectedStatus
	if expectedStatus == 0 {
		expectedStatus = 200
	}

	if resp.StatusCode == expectedStatus {
		result.Healthy = true
		result.Details = fmt.Sprintf("HTTP %d", resp.StatusCode)
	} else {
		result.Error = fmt.Sprintf("expected %d, got %d", expectedStatus, resp.StatusCode)
	}

	return result
}

func (u *DepsUsecase) checkTCP(ctx context.Context, dep models.Dependency, result models.DependencyResult) models.DependencyResult {
	address := fmt.Sprintf("%s:%d", dep.Host, dep.Port)
	dialer := &net.Dialer{Timeout: 10 * time.Second}

	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer conn.Close()

	result.Healthy = true
	result.Details = fmt.Sprintf("Connected to %s", address)
	return result
}

func (u *DepsUsecase) checkDNS(ctx context.Context, dep models.Dependency, result models.DependencyResult) models.DependencyResult {
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: 5 * time.Second}
			server := dep.Host
			if dep.Port > 0 {
				server = fmt.Sprintf("%s:%d", dep.Host, dep.Port)
			} else {
				server = dep.Host + ":53"
			}
			return d.DialContext(ctx, "udp", server)
		},
	}

	// Test with google.com
	ips, err := resolver.LookupIP(ctx, "ip4", "google.com")
	if err != nil {
		result.Error = err.Error()
		return result
	}

	result.Healthy = true
	result.Details = fmt.Sprintf("Resolved to %d IPs", len(ips))
	return result
}

// CreateQuickDeps creates a quick dependency list from common args
func (u *DepsUsecase) CreateQuickDeps(urls []string, tcpAddrs []string) *models.DepsConfig {
	config := &models.DepsConfig{
		Name:         "Quick Check",
		Dependencies: []models.Dependency{},
	}

	for i, url := range urls {
		config.Dependencies = append(config.Dependencies, models.Dependency{
			Name:     fmt.Sprintf("HTTP-%d", i+1),
			Type:     models.DepTypeHTTP,
			URL:      url,
			Required: true,
		})
	}

	for _, addr := range tcpAddrs {
		var host string
		var port int
		fmt.Sscanf(addr, "%s:%d", &host, &port)
		if host == "" {
			parts := splitHostPort(addr)
			host = parts[0]
			if len(parts) > 1 {
				fmt.Sscanf(parts[1], "%d", &port)
			}
		}
		if port == 0 {
			port = 80
		}

		config.Dependencies = append(config.Dependencies, models.Dependency{
			Name:     fmt.Sprintf("TCP-%s:%d", host, port),
			Type:     models.DepTypeTCP,
			Host:     host,
			Port:     port,
			Required: true,
		})
	}

	return config
}

func splitHostPort(addr string) []string {
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return []string{addr[:idx], addr[idx+1:]}
	}
	return []string{addr}
}
