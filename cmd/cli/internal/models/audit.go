package models

import "time"

// AuditConfig represents the configuration for a full audit
type AuditConfig struct {
	Target        string        `json:"target"`
	Timeout       time.Duration `json:"timeout"`
	SkipMTR       bool          `json:"skip_mtr"`
	SkipPorts     bool          `json:"skip_ports"`
	OutputFormat  string        `json:"output_format"` // terminal, json, markdown
	SaveToFile    string        `json:"save_to_file"`
}

// AuditReport represents a complete audit report
type AuditReport struct {
	// Metadata
	Target      string    `json:"target"`
	Timestamp   time.Time `json:"timestamp"`
	Duration    time.Duration `json:"duration"`

	// Overall Score
	OverallScore    int           `json:"overall_score"`    // 0-100
	OverallGrade    string        `json:"overall_grade"`    // A+ to F
	OverallStatus   AuditStatus   `json:"overall_status"`

	// Section Results
	Connectivity    *AuditConnectivity    `json:"connectivity"`
	SSL             *AuditSSL             `json:"ssl"`
	SecurityHeaders *AuditSecurityHeaders `json:"security_headers"`
	Performance     *AuditPerformance     `json:"performance"`
	NetworkPath     *AuditNetworkPath     `json:"network_path,omitempty"`
	PortScan        *AuditPortScan        `json:"port_scan,omitempty"`

	// Summary
	Issues       []AuditIssue      `json:"issues"`
	Warnings     []AuditWarning    `json:"warnings"`
	Recommendations []string       `json:"recommendations"`
}

// AuditStatus represents the overall audit status
type AuditStatus string

const (
	AuditStatusHealthy  AuditStatus = "HEALTHY"
	AuditStatusWarning  AuditStatus = "WARNING"
	AuditStatusCritical AuditStatus = "CRITICAL"
	AuditStatusDown     AuditStatus = "DOWN"
)

// AuditConnectivity contains connectivity test results
type AuditConnectivity struct {
	Score       int           `json:"score"`
	DNSResolves bool          `json:"dns_resolves"`
	DNSTime     time.Duration `json:"dns_time"`
	IPs         []string      `json:"ips"`
	TCPConnects bool          `json:"tcp_connects"`
	TCPTime     time.Duration `json:"tcp_time"`
	HTTPWorks   bool          `json:"http_works"`
	HTTPStatus  int           `json:"http_status"`
	HTTPTime    time.Duration `json:"http_time"`
}

// AuditSSL contains SSL/TLS analysis results
type AuditSSL struct {
	Score           int           `json:"score"`
	Valid           bool          `json:"valid"`
	Version         string        `json:"version"`
	CipherSuite     string        `json:"cipher_suite"`
	Certificate     string        `json:"certificate"`
	Issuer          string        `json:"issuer"`
	ExpiresAt       time.Time     `json:"expires_at"`
	DaysUntilExpiry int           `json:"days_until_expiry"`
	Chain           []string      `json:"chain"`
	Errors          []string      `json:"errors,omitempty"`
}

// AuditSecurityHeaders contains security headers analysis
type AuditSecurityHeaders struct {
	Score       int                     `json:"score"`
	Grade       string                  `json:"grade"`
	Headers     []AuditHeaderCheck      `json:"headers"`
	Missing     []string                `json:"missing"`
}

// AuditHeaderCheck represents a single header check
type AuditHeaderCheck struct {
	Name        string `json:"name"`
	Present     bool   `json:"present"`
	Value       string `json:"value,omitempty"`
	Secure      bool   `json:"secure"`
	Description string `json:"description"`
}

// AuditPerformance contains performance metrics
type AuditPerformance struct {
	Score              int           `json:"score"`
	DNSLookup          time.Duration `json:"dns_lookup"`
	TCPConnect         time.Duration `json:"tcp_connect"`
	TLSHandshake       time.Duration `json:"tls_handshake"`
	ServerProcessing   time.Duration `json:"server_processing"`
	ContentTransfer    time.Duration `json:"content_transfer"`
	TotalTime          time.Duration `json:"total_time"`
	TTFB               time.Duration `json:"ttfb"` // Time to first byte
	Rating             string        `json:"rating"` // Fast, Normal, Slow
}

// AuditNetworkPath contains network path analysis
type AuditNetworkPath struct {
	Score         int           `json:"score"`
	TotalHops     int           `json:"total_hops"`
	Completed     bool          `json:"completed"`
	TotalLatency  time.Duration `json:"total_latency"`
	Bottleneck    string        `json:"bottleneck,omitempty"`
	Hops          []AuditHop    `json:"hops"`
}

// AuditHop represents a single hop in the network path
type AuditHop struct {
	Number     int           `json:"number"`
	Host       string        `json:"host"`
	Latency    time.Duration `json:"latency"`
	Loss       float64       `json:"loss"`
}

// AuditPortScan contains port scan results
type AuditPortScan struct {
	Score      int          `json:"score"`
	OpenPorts  []AuditPort  `json:"open_ports"`
	TotalScanned int        `json:"total_scanned"`
}

// AuditPort represents a scanned port
type AuditPort struct {
	Port    int    `json:"port"`
	State   string `json:"state"`
	Service string `json:"service"`
}

// AuditIssue represents a critical issue found
type AuditIssue struct {
	Severity    string `json:"severity"` // critical, high, medium, low
	Category    string `json:"category"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Solution    string `json:"solution,omitempty"`
}

// AuditWarning represents a warning
type AuditWarning struct {
	Category    string `json:"category"`
	Message     string `json:"message"`
	Suggestion  string `json:"suggestion,omitempty"`
}
