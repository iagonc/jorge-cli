package models

import "time"

// DiagnoseResult represents the complete diagnosis of a target
type DiagnoseResult struct {
	Target      string              `json:"target"`
	StartTime   time.Time           `json:"start_time"`
	Duration    time.Duration       `json:"duration"`
	Overall     HealthStatus        `json:"overall_status"`
	Summary     string              `json:"summary"`
	Checks      []DiagnoseCheck     `json:"checks"`
	Problems    []Problem           `json:"problems"`
	Suggestions []Suggestion        `json:"suggestions"`
}

// DiagnoseCheck represents a single diagnostic check
type DiagnoseCheck struct {
	Name        string        `json:"name"`
	Category    CheckCategory `json:"category"`
	Status      CheckStatus   `json:"status"`
	Message     string        `json:"message"`
	Details     string        `json:"details,omitempty"`
	Duration    time.Duration `json:"duration"`
	RawData     interface{}   `json:"raw_data,omitempty"`
}

// CheckCategory represents the category of a check
type CheckCategory string

const (
	CategoryDNS          CheckCategory = "DNS"
	CategoryConnectivity CheckCategory = "Conectividade"
	CategorySSL          CheckCategory = "SSL/TLS"
	CategoryHTTP         CheckCategory = "HTTP"
	CategoryLatency      CheckCategory = "Latência"
	CategoryFirewall     CheckCategory = "Firewall"
)

// CheckStatus represents the status of a check
type CheckStatus string

const (
	CheckPassed  CheckStatus = "passed"
	CheckWarning CheckStatus = "warning"
	CheckFailed  CheckStatus = "failed"
	CheckSkipped CheckStatus = "skipped"
)

// Problem represents an identified problem
type Problem struct {
	Severity    Severity `json:"severity"`
	Category    CheckCategory `json:"category"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	TechDetails string   `json:"technical_details,omitempty"`
}

// Severity represents problem severity
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityWarning  Severity = "warning"
	SeverityInfo     Severity = "info"
)

// Suggestion represents a fix suggestion
type Suggestion struct {
	Priority    int    `json:"priority"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Command     string `json:"command,omitempty"`
	Link        string `json:"link,omitempty"`
}

// DoctorResult represents local network health check
type DoctorResult struct {
	Timestamp    time.Time       `json:"timestamp"`
	Duration     time.Duration   `json:"duration"`
	Overall      HealthStatus    `json:"overall_status"`
	Connectivity ConnectivityInfo `json:"connectivity"`
	DNS          DNSInfo          `json:"dns"`
	Gateway      GatewayInfo      `json:"gateway"`
	Internet     InternetInfo     `json:"internet"`
	Problems     []Problem        `json:"problems"`
	Suggestions  []Suggestion     `json:"suggestions"`
}

// ConnectivityInfo represents local connectivity info
type ConnectivityInfo struct {
	HasIPv4      bool     `json:"has_ipv4"`
	HasIPv6      bool     `json:"has_ipv6"`
	LocalIPs     []string `json:"local_ips"`
	Interfaces   []string `json:"interfaces"`
}

// DNSInfo represents DNS configuration info
type DNSInfo struct {
	Servers      []string `json:"servers"`
	CanResolve   bool     `json:"can_resolve"`
	ResponseTime time.Duration `json:"response_time"`
}

// GatewayInfo represents gateway info
type GatewayInfo struct {
	IP           string        `json:"ip"`
	Reachable    bool          `json:"reachable"`
	ResponseTime time.Duration `json:"response_time"`
}

// InternetInfo represents internet connectivity info
type InternetInfo struct {
	Connected    bool          `json:"connected"`
	PublicIP     string        `json:"public_ip,omitempty"`
	ResponseTime time.Duration `json:"response_time"`
}
