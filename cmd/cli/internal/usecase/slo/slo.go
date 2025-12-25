package slo

import (
	"fmt"
	"os"
	"time"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// SLOUsecase handles SLO operations
type SLOUsecase struct {
	logger *zap.Logger
}

// NewSLOUsecase creates a new SLOUsecase
func NewSLOUsecase(logger *zap.Logger) *SLOUsecase {
	return &SLOUsecase{logger: logger}
}

// LoadConfig loads SLO configuration from a file
func (u *SLOUsecase) LoadConfig(path string) (*models.SLOConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config models.SLOConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// CalculateBudget calculates error budget for an SLO
func (u *SLOUsecase) CalculateBudget(slo models.SLO, currentSLI float64) models.ErrorBudget {
	budget := models.ErrorBudget{
		Window:       slo.Window,
		Target:       slo.Target,
		Current:      currentSLI,
		CalculatedAt: time.Now(),
	}

	// Calculate total budget (100 - target)
	budget.BudgetTotal = 100 - slo.Target

	// Calculate used budget (target - current, if current < target)
	if currentSLI < slo.Target {
		budget.BudgetUsed = slo.Target - currentSLI
	} else {
		budget.BudgetUsed = 0
	}

	// Calculate remaining
	budget.BudgetRemaining = budget.BudgetTotal - budget.BudgetUsed
	if budget.BudgetRemaining < 0 {
		budget.BudgetRemaining = 0
	}

	// Calculate percentage remaining
	if budget.BudgetTotal > 0 {
		budget.BudgetPercent = (budget.BudgetRemaining / budget.BudgetTotal) * 100
	}

	// Determine status
	switch {
	case budget.BudgetRemaining <= 0:
		budget.Status = "exhausted"
	case budget.BudgetPercent < 25:
		budget.Status = "critical"
	case budget.BudgetPercent < 50:
		budget.Status = "warning"
	default:
		budget.Status = "healthy"
	}

	return budget
}

// GenerateReport generates an SLO report
func (u *SLOUsecase) GenerateReport(slo models.SLO, goodEvents, totalEvents int64, period string) models.SLOReport {
	report := models.SLOReport{
		SLO:         slo,
		Period:      period,
		StartTime:   time.Now().Add(-u.parseDuration(period)),
		EndTime:     time.Now(),
		TotalEvents: totalEvents,
		GoodEvents:  goodEvents,
		BadEvents:   totalEvents - goodEvents,
		Target:      slo.Target,
	}

	// Calculate current SLI
	if totalEvents > 0 {
		report.CurrentSLI = (float64(goodEvents) / float64(totalEvents)) * 100
	}

	// Check compliance
	report.Compliance = report.CurrentSLI >= slo.Target

	// Calculate error budget
	report.ErrorBudget = u.CalculateBudget(slo, report.CurrentSLI)

	return report
}

// CalculateDowntimeAllowance calculates allowed downtime based on SLO
func (u *SLOUsecase) CalculateDowntimeAllowance(target float64, window string) map[string]string {
	windowDuration := u.parseDuration(window)
	allowedDowntimePercent := 100 - target
	// Use float64 for precise calculation, then convert to duration
	allowedDowntimeNanos := float64(windowDuration) * allowedDowntimePercent / 100
	allowedDowntime := time.Duration(allowedDowntimeNanos)

	result := make(map[string]string)
	result["window"] = window
	result["target"] = fmt.Sprintf("%.2f%%", target)
	result["allowed_downtime_percent"] = fmt.Sprintf("%.3f%%", allowedDowntimePercent)

	// Convert to human readable
	if allowedDowntime >= 24*time.Hour {
		result["allowed_downtime"] = fmt.Sprintf("%.2f days", allowedDowntime.Hours()/24)
	} else if allowedDowntime >= time.Hour {
		result["allowed_downtime"] = fmt.Sprintf("%.2f hours", allowedDowntime.Hours())
	} else if allowedDowntime >= time.Minute {
		result["allowed_downtime"] = fmt.Sprintf("%.2f minutes", allowedDowntime.Minutes())
	} else {
		result["allowed_downtime"] = fmt.Sprintf("%.2f seconds", allowedDowntime.Seconds())
	}

	// Per month/week/day breakdown - calculate based on window in days
	windowDays := float64(windowDuration) / float64(24*time.Hour)
	monthlyDowntime := time.Duration(float64(allowedDowntime) * 30 / windowDays)
	weeklyDowntime := time.Duration(float64(allowedDowntime) * 7 / windowDays)
	dailyDowntime := time.Duration(float64(allowedDowntime) / windowDays)

	result["monthly_allowance"] = formatDuration(monthlyDowntime)
	result["weekly_allowance"] = formatDuration(weeklyDowntime)
	result["daily_allowance"] = formatDuration(dailyDowntime)

	return result
}

// GetDashboard returns a dashboard view of SLOs
func (u *SLOUsecase) GetDashboard(config *models.SLOConfig) *models.SLODashboard {
	dashboard := &models.SLODashboard{
		Service:     config.Service,
		SLOs:        []models.SLOReport{},
		LastUpdated: time.Now(),
	}

	worstStatus := "healthy"

	for _, slo := range config.SLOs {
		// For demo purposes, generate sample data
		// In production, this would query Prometheus/DataDog/etc
		totalEvents := int64(100000)
		goodEvents := int64(float64(totalEvents) * (slo.Target + 0.5) / 100)

		report := u.GenerateReport(slo, goodEvents, totalEvents, slo.Window)
		dashboard.SLOs = append(dashboard.SLOs, report)

		// Track worst status
		switch report.ErrorBudget.Status {
		case "exhausted":
			worstStatus = "critical"
		case "critical":
			if worstStatus != "critical" {
				worstStatus = "critical"
			}
		case "warning":
			if worstStatus == "healthy" {
				worstStatus = "warning"
			}
		}
	}

	dashboard.OverallHealth = worstStatus
	return dashboard
}

// GenerateConfig generates a sample SLO configuration
func (u *SLOUsecase) GenerateConfig() string {
	return `version: "1.0"
service: "my-api"

slos:
  - name: "Availability"
    description: "Service should be available 99.9% of the time"
    service: "my-api"
    sli:
      type: "availability"
      good_events: "http_requests_total{status=~\"2..|3..\"}"
      total_events: "http_requests_total"
    target: 99.9
    window: "30d"

  - name: "Latency P99"
    description: "99th percentile latency should be under 200ms"
    service: "my-api"
    sli:
      type: "latency"
      threshold: "200ms"
      query: "histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))"
    target: 99.0
    window: "30d"

  - name: "Error Rate"
    description: "Error rate should be below 0.1%"
    service: "my-api"
    sli:
      type: "error_rate"
      good_events: "http_requests_total{status!~\"5..\"}"
      total_events: "http_requests_total"
    target: 99.9
    window: "7d"
`
}

func (u *SLOUsecase) parseDuration(s string) time.Duration {
	// Parse custom duration format (30d, 7d, etc)
	if len(s) == 0 {
		return 30 * 24 * time.Hour
	}

	var value int
	var unit string
	fmt.Sscanf(s, "%d%s", &value, &unit)

	switch unit {
	case "d":
		return time.Duration(value) * 24 * time.Hour
	case "h":
		return time.Duration(value) * time.Hour
	case "m":
		return time.Duration(value) * time.Minute
	case "w":
		return time.Duration(value) * 7 * 24 * time.Hour
	default:
		return 30 * 24 * time.Hour
	}
}

func formatDuration(d time.Duration) string {
	if d >= 24*time.Hour {
		return fmt.Sprintf("%.2fh (%.0fd)", d.Hours(), d.Hours()/24)
	}
	if d >= time.Hour {
		return fmt.Sprintf("%.2fh", d.Hours())
	}
	if d >= time.Minute {
		return fmt.Sprintf("%.2fm", d.Minutes())
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}
