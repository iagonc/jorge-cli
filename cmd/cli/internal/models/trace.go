package models

import "time"

// TraceResult represents the complete trace of an HTTP request
type TraceResult struct {
	URL           string            `json:"url"`
	Method        string            `json:"method"`
	Phases        []TracePhase      `json:"phases"`
	TotalDuration time.Duration     `json:"total_duration"`
	Success       bool              `json:"success"`
	StatusCode    int               `json:"status_code,omitempty"`
	Error         string            `json:"error,omitempty"`
	Response      *TraceResponse    `json:"response,omitempty"`
	TLS           *TraceTLSInfo     `json:"tls,omitempty"`
	Redirects     []TraceRedirect   `json:"redirects,omitempty"`
}

// TracePhase represents a single phase in the request lifecycle
type TracePhase struct {
	Name      string        `json:"name"`
	Duration  time.Duration `json:"duration"`
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	Details   string        `json:"details,omitempty"`
	Success   bool          `json:"success"`
	Error     string        `json:"error,omitempty"`
}

// TraceResponse contains response details
type TraceResponse struct {
	StatusCode    int               `json:"status_code"`
	Status        string            `json:"status"`
	Headers       map[string]string `json:"headers"`
	ContentType   string            `json:"content_type"`
	ContentLength int64             `json:"content_length"`
	Body          string            `json:"body,omitempty"`
}

// TraceTLSInfo contains TLS handshake details
type TraceTLSInfo struct {
	Version            string    `json:"version"`
	CipherSuite        string    `json:"cipher_suite"`
	ServerName         string    `json:"server_name"`
	CertificateSubject string    `json:"certificate_subject"`
	CertificateIssuer  string    `json:"certificate_issuer"`
	CertificateExpiry  time.Time `json:"certificate_expiry"`
	HandshakeDuration  time.Duration `json:"handshake_duration"`
}

// TraceRedirect represents a redirect in the chain
type TraceRedirect struct {
	From       string `json:"from"`
	To         string `json:"to"`
	StatusCode int    `json:"status_code"`
}
