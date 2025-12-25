package models

import "time"

// MockConfig represents the configuration for a mock server
type MockConfig struct {
	Port          int             `json:"port"`
	Host          string          `json:"host"`
	Routes        []MockRoute     `json:"routes,omitempty"`
	DefaultStatus int             `json:"default_status"`
	LogRequests   bool            `json:"log_requests"`
	CORS          bool            `json:"cors"`
	Delay         time.Duration   `json:"delay,omitempty"`
}

// MockRoute represents a configured route
type MockRoute struct {
	Method      string            `json:"method"`
	Path        string            `json:"path"`
	StatusCode  int               `json:"status_code"`
	Body        string            `json:"body,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Delay       time.Duration     `json:"delay,omitempty"`
	ContentType string            `json:"content_type,omitempty"`
}

// MockRequest represents a received request
type MockRequest struct {
	ID          string            `json:"id"`
	Timestamp   time.Time         `json:"timestamp"`
	Method      string            `json:"method"`
	Path        string            `json:"path"`
	Query       string            `json:"query,omitempty"`
	Headers     map[string]string `json:"headers"`
	Body        string            `json:"body,omitempty"`
	RemoteAddr  string            `json:"remote_addr"`
	ContentType string            `json:"content_type,omitempty"`
	BodySize    int64             `json:"body_size"`
}

// MockResponse represents the response sent
type MockResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body,omitempty"`
	Duration   time.Duration     `json:"duration"`
}

// MockStats contains statistics about the mock server
type MockStats struct {
	StartTime       time.Time         `json:"start_time"`
	TotalRequests   int               `json:"total_requests"`
	RequestsByPath  map[string]int    `json:"requests_by_path"`
	RequestsByMethod map[string]int   `json:"requests_by_method"`
	AverageLatency  time.Duration     `json:"average_latency"`
}
