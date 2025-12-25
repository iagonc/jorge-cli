package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// TestConfig represents a configuration for testing purposes
type TestConfig struct {
	APIBaseURL string
	Timeout    time.Duration
	Version    string
}

// MockHTTPResponse creates a mock HTTP response for testing
func MockHTTPResponse(statusCode int, body interface{}) *http.Response {
	var bodyReader io.ReadCloser

	switch v := body.(type) {
	case string:
		bodyReader = io.NopCloser(strings.NewReader(v))
	case []byte:
		bodyReader = io.NopCloser(bytes.NewReader(v))
	default:
		jsonBody, _ := json.Marshal(body)
		bodyReader = io.NopCloser(bytes.NewReader(jsonBody))
	}

	return &http.Response{
		StatusCode: statusCode,
		Body:       bodyReader,
		Header:     make(http.Header),
	}
}

// NewTestLogger creates a no-op logger for testing
func NewTestLogger() *zap.Logger {
	return zap.NewNop()
}

// NewTestConfig creates a test configuration
func NewTestConfig(baseURL string) *TestConfig {
	return &TestConfig{
		APIBaseURL: baseURL,
		Timeout:    10 * time.Second,
		Version:    "test",
	}
}
