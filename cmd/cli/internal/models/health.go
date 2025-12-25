package models

import "time"

// HealthCheckConfig represents the YAML configuration for health checks
type HealthCheckConfig struct {
	Checks []HealthTarget `yaml:"checks" json:"checks"`
}

// HealthTarget represents a single health check target
type HealthTarget struct {
	Name            string            `yaml:"name" json:"name"`
	URL             string            `yaml:"url" json:"url"`
	Method          string            `yaml:"method" json:"method"`
	Headers         map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`
	Body            string            `yaml:"body,omitempty" json:"body,omitempty"`
	ExpectedStatus  int               `yaml:"expected_status" json:"expected_status"`
	ExpectedBody    string            `yaml:"expected_body,omitempty" json:"expected_body,omitempty"`
	Timeout         int               `yaml:"timeout" json:"timeout"` // seconds
	FollowRedirects bool              `yaml:"follow_redirects" json:"follow_redirects"`
}

// HealthStatus represents the health status
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusUnhealthy HealthStatus = "unhealthy"
	StatusError     HealthStatus = "error"
)

// HealthCheckResult represents the result of a single health check
type HealthCheckResult struct {
	Name         string        `json:"name"`
	URL          string        `json:"url"`
	Status       HealthStatus  `json:"status"`
	StatusCode   int           `json:"status_code"`
	ResponseTime time.Duration `json:"response_time"`
	BodyMatch    bool          `json:"body_match"`
	Error        string        `json:"error,omitempty"`
	CheckedAt    time.Time     `json:"checked_at"`
}

// HealthCheckSummary contains the summary of all health checks
type HealthCheckSummary struct {
	TotalChecks int                 `json:"total_checks"`
	Healthy     int                 `json:"healthy"`
	Unhealthy   int                 `json:"unhealthy"`
	Errors      int                 `json:"errors"`
	Results     []HealthCheckResult `json:"results"`
	Duration    time.Duration       `json:"duration"`
}
