package models

import "time"

// CurlRequest represents an HTTP request configuration
type CurlRequest struct {
	URL             string            `json:"url"`
	Method          string            `json:"method"`
	Headers         map[string]string `json:"headers,omitempty"`
	Body            string            `json:"body,omitempty"`
	Timeout         time.Duration     `json:"timeout"`
	FollowRedirects bool              `json:"follow_redirects"`
	Insecure        bool              `json:"insecure"`
	Verbose         bool              `json:"verbose"`
	ShowBody        bool              `json:"show_body"`
	MaxBodySize     int64             `json:"max_body_size"`
}

// CurlResult represents the result of an HTTP request
type CurlResult struct {
	Request      CurlRequestInfo  `json:"request"`
	Response     CurlResponseInfo `json:"response"`
	Timing       CurlTiming       `json:"timing"`
	TLS          *CurlTLSInfo     `json:"tls,omitempty"`
	Redirects    []CurlRedirect   `json:"redirects,omitempty"`
	Error        string           `json:"error,omitempty"`
	Success      bool             `json:"success"`
}

// CurlRequestInfo contains request details
type CurlRequestInfo struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body,omitempty"`
}

// CurlResponseInfo contains response details
type CurlResponseInfo struct {
	StatusCode    int               `json:"status_code"`
	Status        string            `json:"status"`
	Headers       map[string]string `json:"headers"`
	Body          string            `json:"body,omitempty"`
	BodySize      int64             `json:"body_size"`
	ContentType   string            `json:"content_type"`
	Cookies       []CurlCookie      `json:"cookies,omitempty"`
}

// CurlCookie represents a cookie from the response
type CurlCookie struct {
	Name     string    `json:"name"`
	Value    string    `json:"value"`
	Path     string    `json:"path"`
	Domain   string    `json:"domain"`
	Expires  time.Time `json:"expires,omitempty"`
	Secure   bool      `json:"secure"`
	HttpOnly bool      `json:"http_only"`
}

// CurlTiming contains timing information
type CurlTiming struct {
	DNSLookup        time.Duration `json:"dns_lookup"`
	TCPConnect       time.Duration `json:"tcp_connect"`
	TLSHandshake     time.Duration `json:"tls_handshake,omitempty"`
	ServerProcessing time.Duration `json:"server_processing"`
	ContentTransfer  time.Duration `json:"content_transfer"`
	Total            time.Duration `json:"total"`
}

// CurlTLSInfo contains TLS details
type CurlTLSInfo struct {
	Version     string `json:"version"`
	CipherSuite string `json:"cipher_suite"`
	Certificate string `json:"certificate"`
	Issuer      string `json:"issuer"`
}

// CurlRedirect represents a redirect
type CurlRedirect struct {
	From       string `json:"from"`
	To         string `json:"to"`
	StatusCode int    `json:"status_code"`
}
