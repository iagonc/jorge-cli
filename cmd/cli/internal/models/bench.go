package models

import "time"

// BenchmarkConfig contains the configuration for HTTP benchmarking
type BenchmarkConfig struct {
	URL         string            `json:"url"`
	Method      string            `json:"method"`
	Headers     map[string]string `json:"headers,omitempty"`
	Body        string            `json:"body,omitempty"`
	Requests    int               `json:"requests"`
	Concurrency int               `json:"concurrency"`
	Duration    time.Duration     `json:"duration,omitempty"` // Alternative to fixed request count
	Timeout     time.Duration     `json:"timeout"`
}

// LatencyStats contains latency statistics
type LatencyStats struct {
	Min  time.Duration `json:"min"`
	Max  time.Duration `json:"max"`
	Mean time.Duration `json:"mean"`
	P50  time.Duration `json:"p50"`
	P95  time.Duration `json:"p95"`
	P99  time.Duration `json:"p99"`
}

// BenchmarkResult contains the complete benchmark result
type BenchmarkResult struct {
	URL               string         `json:"url"`
	TotalRequests     int            `json:"total_requests"`
	SuccessfulRequests int           `json:"successful_requests"`
	FailedRequests    int            `json:"failed_requests"`
	TotalDuration     time.Duration  `json:"total_duration"`
	RequestsPerSecond float64        `json:"requests_per_second"`
	BytesReceived     int64          `json:"bytes_received"`
	Latency           LatencyStats   `json:"latency"`
	StatusCodes       map[int]int    `json:"status_codes"`
	Errors            map[string]int `json:"errors,omitempty"`
}

// RequestResult contains the result of a single request
type RequestResult struct {
	Duration   time.Duration `json:"duration"`
	StatusCode int           `json:"status_code"`
	Error      error         `json:"-"`
	Size       int64         `json:"size"`
}
