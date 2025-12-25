package status

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// StatusUsecase handles status page operations
type StatusUsecase struct {
	logger *zap.Logger
}

// NewStatusUsecase creates a new StatusUsecase
func NewStatusUsecase(logger *zap.Logger) *StatusUsecase {
	return &StatusUsecase{logger: logger}
}

// LoadConfig loads a status page configuration from a YAML file
func (u *StatusUsecase) LoadConfig(path string) (*models.StatusPageConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config models.StatusPageConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// CheckAll checks all components and returns the status page result
func (u *StatusUsecase) CheckAll(ctx context.Context, config *models.StatusPageConfig) (*models.StatusPageResult, error) {
	result := &models.StatusPageResult{
		Title:       config.Title,
		Description: config.Description,
		Components:  make([]models.StatusComponent, len(config.Components)),
		LastUpdated: time.Now(),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, comp := range config.Components {
		wg.Add(1)
		go func(idx int, component models.StatusComponent) {
			defer wg.Done()

			status := u.checkComponent(ctx, component)

			mu.Lock()
			result.Components[idx] = component
			result.Components[idx].Status = status
			mu.Unlock()
		}(i, comp)
	}

	wg.Wait()

	// Calculate overall status
	result.OverallStatus = u.calculateOverallStatus(result.Components)

	// Filter active incidents
	for _, incident := range config.Incidents {
		if incident.ResolvedAt == nil {
			result.ActiveIncidents = append(result.ActiveIncidents, incident)
		}
	}

	return result, nil
}

func (u *StatusUsecase) checkComponent(ctx context.Context, comp models.StatusComponent) models.ComponentStatus {
	status := models.ComponentStatus{
		LastChecked: time.Now(),
		State:       models.StatusOperational,
	}

	start := time.Now()

	switch comp.Type {
	case "http", "https":
		status = u.checkHTTP(ctx, comp.Target)
	case "tcp":
		status = u.checkTCP(ctx, comp.Target)
	case "dns":
		status = u.checkDNS(ctx, comp.Target)
	default:
		status.State = models.StatusUnknown
		status.Message = "Unknown check type"
	}

	status.Latency = time.Since(start).Round(time.Millisecond).String()
	status.LastChecked = time.Now()

	return status
}

func (u *StatusUsecase) checkHTTP(ctx context.Context, target string) models.ComponentStatus {
	status := models.ComponentStatus{}

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequestWithContext(ctx, "GET", target, nil)
	if err != nil {
		status.State = models.StatusMajorOutage
		status.Message = err.Error()
		return status
	}

	resp, err := client.Do(req)
	if err != nil {
		status.State = models.StatusMajorOutage
		status.Message = err.Error()
		return status
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		status.State = models.StatusOperational
		status.Message = fmt.Sprintf("HTTP %d", resp.StatusCode)
	case resp.StatusCode >= 300 && resp.StatusCode < 400:
		status.State = models.StatusOperational
		status.Message = fmt.Sprintf("HTTP %d (redirect)", resp.StatusCode)
	case resp.StatusCode >= 400 && resp.StatusCode < 500:
		status.State = models.StatusDegraded
		status.Message = fmt.Sprintf("HTTP %d", resp.StatusCode)
	case resp.StatusCode >= 500:
		status.State = models.StatusMajorOutage
		status.Message = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}

	return status
}

func (u *StatusUsecase) checkTCP(ctx context.Context, target string) models.ComponentStatus {
	status := models.ComponentStatus{}

	dialer := &net.Dialer{Timeout: 5 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", target)
	if err != nil {
		status.State = models.StatusMajorOutage
		status.Message = err.Error()
		return status
	}
	conn.Close()

	status.State = models.StatusOperational
	status.Message = "Connection successful"
	return status
}

func (u *StatusUsecase) checkDNS(ctx context.Context, target string) models.ComponentStatus {
	status := models.ComponentStatus{}

	resolver := &net.Resolver{
		PreferGo: true,
	}

	ips, err := resolver.LookupIPAddr(ctx, target)
	if err != nil {
		status.State = models.StatusMajorOutage
		status.Message = err.Error()
		return status
	}

	if len(ips) == 0 {
		status.State = models.StatusMajorOutage
		status.Message = "No IP addresses found"
		return status
	}

	status.State = models.StatusOperational
	status.Message = fmt.Sprintf("Resolved to %d IPs", len(ips))
	return status
}

func (u *StatusUsecase) calculateOverallStatus(components []models.StatusComponent) models.StatusState {
	if len(components) == 0 {
		return models.StatusUnknown
	}

	var operational, degraded, outage int

	for _, comp := range components {
		switch comp.Status.State {
		case models.StatusOperational:
			operational++
		case models.StatusDegraded:
			degraded++
		case models.StatusPartialOutage, models.StatusMajorOutage:
			outage++
		}
	}

	total := len(components)

	if outage == total {
		return models.StatusMajorOutage
	}
	if outage > 0 {
		return models.StatusPartialOutage
	}
	if degraded > 0 {
		return models.StatusDegraded
	}
	return models.StatusOperational
}

// CheckSingle checks a single target
func (u *StatusUsecase) CheckSingle(ctx context.Context, target, checkType string) models.ComponentStatus {
	comp := models.StatusComponent{
		Name:   target,
		Type:   checkType,
		Target: target,
	}
	return u.checkComponent(ctx, comp)
}

// GenerateConfig generates a sample status page configuration
func (u *StatusUsecase) GenerateConfig() string {
	sample := `title: "My Service Status"
description: "Current status of all services"

components:
  - name: "API"
    description: "Main API endpoint"
    type: "http"
    target: "https://api.example.com/health"
    group: "Backend"

  - name: "Website"
    description: "Public website"
    type: "http"
    target: "https://example.com"
    group: "Frontend"

  - name: "Database"
    description: "Primary database"
    type: "tcp"
    target: "db.internal:5432"
    group: "Infrastructure"

  - name: "Cache"
    description: "Redis cache"
    type: "tcp"
    target: "redis.internal:6379"
    group: "Infrastructure"

  - name: "DNS"
    description: "DNS resolution"
    type: "dns"
    target: "example.com"
    group: "Infrastructure"

incidents: []
`
	return sample
}
