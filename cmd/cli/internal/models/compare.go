package models

import "time"

// CompareResult represents the comparison between two endpoints
type CompareResult struct {
	Endpoint1   EndpointResult    `json:"endpoint1"`
	Endpoint2   EndpointResult    `json:"endpoint2"`
	Differences []CompareDiff     `json:"differences"`
	Summary     string            `json:"summary"`
	Duration    time.Duration     `json:"duration"`
}

// EndpointResult represents the result for a single endpoint
type EndpointResult struct {
	Target       string            `json:"target"`
	Reachable    bool              `json:"reachable"`
	DNSTime      time.Duration     `json:"dns_time"`
	ConnectTime  time.Duration     `json:"connect_time"`
	TLSTime      time.Duration     `json:"tls_time,omitempty"`
	ResponseTime time.Duration     `json:"response_time,omitempty"`
	StatusCode   int               `json:"status_code,omitempty"`
	TLSVersion   string            `json:"tls_version,omitempty"`
	IP           string            `json:"ip,omitempty"`
	Headers      map[string]string `json:"headers,omitempty"`
	Error        string            `json:"error,omitempty"`
}

// CompareDiff represents a difference between two endpoints
type CompareDiff struct {
	Field       string `json:"field"`
	Value1      string `json:"value1"`
	Value2      string `json:"value2"`
	Severity    string `json:"severity"` // info, warning, critical
	Description string `json:"description"`
}
