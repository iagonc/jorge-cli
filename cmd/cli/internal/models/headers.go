package models

import "time"

// HeadersResult represents the security headers analysis
type HeadersResult struct {
	URL           string               `json:"url"`
	StatusCode    int                  `json:"status_code"`
	Headers       map[string]string    `json:"headers"`
	SecurityScore int                  `json:"security_score"` // 0-100
	Grade         string               `json:"grade"`          // A+, A, B, C, D, F
	Checks        []SecurityHeaderCheck `json:"checks"`
	Duration      time.Duration        `json:"duration"`
	Suggestions   []string             `json:"suggestions,omitempty"`
}

// SecurityHeaderCheck represents a single security header check
type SecurityHeaderCheck struct {
	Name        string            `json:"name"`
	Header      string            `json:"header"`
	Present     bool              `json:"present"`
	Value       string            `json:"value,omitempty"`
	Status      HeaderCheckStatus `json:"status"`
	Description string            `json:"description"`
	Impact      string            `json:"impact"`
	Suggestion  string            `json:"suggestion,omitempty"`
}

// HeaderCheckStatus represents the status of a header check
type HeaderCheckStatus string

const (
	HeaderStatusPass    HeaderCheckStatus = "pass"
	HeaderStatusWarn    HeaderCheckStatus = "warning"
	HeaderStatusFail    HeaderCheckStatus = "fail"
	HeaderStatusInfo    HeaderCheckStatus = "info"
)

// SecurityHeaders contains common security header names
var SecurityHeaders = map[string]string{
	"Strict-Transport-Security":    "HSTS",
	"Content-Security-Policy":      "CSP",
	"X-Content-Type-Options":       "X-Content-Type-Options",
	"X-Frame-Options":              "X-Frame-Options",
	"X-XSS-Protection":             "X-XSS-Protection",
	"Referrer-Policy":              "Referrer-Policy",
	"Permissions-Policy":           "Permissions-Policy",
	"Access-Control-Allow-Origin":  "CORS",
	"X-Permitted-Cross-Domain-Policies": "Cross-Domain-Policies",
	"Cross-Origin-Opener-Policy":   "COOP",
	"Cross-Origin-Resource-Policy": "CORP",
	"Cross-Origin-Embedder-Policy": "COEP",
}
