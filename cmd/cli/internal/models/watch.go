package models

import "time"

// WatchConfig represents the configuration for watching an endpoint
type WatchConfig struct {
	URL            string            `json:"url"`
	Method         string            `json:"method"`
	Headers        map[string]string `json:"headers,omitempty"`
	Body           string            `json:"body,omitempty"`
	Interval       time.Duration     `json:"interval"`
	Timeout        time.Duration     `json:"timeout"`
	ExpectedStatus int               `json:"expected_status"`
	ExpectedBody   string            `json:"expected_body,omitempty"`
	AlertOnFail    bool              `json:"alert_on_fail"`
	MaxFailures    int               `json:"max_failures"`
}

// WatchEvent represents a single check event
type WatchEvent struct {
	Timestamp    time.Time     `json:"timestamp"`
	Status       WatchStatus   `json:"status"`
	StatusCode   int           `json:"status_code,omitempty"`
	ResponseTime time.Duration `json:"response_time"`
	Error        string        `json:"error,omitempty"`
	BodyMatch    bool          `json:"body_match,omitempty"`
}

// WatchStatus represents the status of a watch check
type WatchStatus string

const (
	WatchStatusUp       WatchStatus = "UP"
	WatchStatusDown     WatchStatus = "DOWN"
	WatchStatusDegraded WatchStatus = "DEGRADED"
)

// WatchStats contains statistics for the watch session
type WatchStats struct {
	TotalChecks      int           `json:"total_checks"`
	SuccessfulChecks int           `json:"successful_checks"`
	FailedChecks     int           `json:"failed_checks"`
	Uptime           float64       `json:"uptime_percent"`
	AvgResponseTime  time.Duration `json:"avg_response_time"`
	MinResponseTime  time.Duration `json:"min_response_time"`
	MaxResponseTime  time.Duration `json:"max_response_time"`
	CurrentStreak    int           `json:"current_streak"` // positive = up, negative = down
	LongestDowntime  time.Duration `json:"longest_downtime"`
}
