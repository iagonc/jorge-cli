package models

import "time"

// TunnelConfig represents the configuration for tunnel/proxy testing
type TunnelConfig struct {
	Target       string        `json:"target"`
	ProxyURL     string        `json:"proxy_url,omitempty"`
	ProxyType    ProxyType     `json:"proxy_type"`
	SSHHost      string        `json:"ssh_host,omitempty"`
	SSHUser      string        `json:"ssh_user,omitempty"`
	SSHKey       string        `json:"ssh_key,omitempty"`
	Timeout      time.Duration `json:"timeout"`
	TestHTTP     bool          `json:"test_http"`
	TestHTTPS    bool          `json:"test_https"`
}

// ProxyType represents the type of proxy
type ProxyType string

const (
	ProxyTypeHTTP    ProxyType = "http"
	ProxyTypeHTTPS   ProxyType = "https"
	ProxyTypeSOCKS4  ProxyType = "socks4"
	ProxyTypeSOCKS5  ProxyType = "socks5"
	ProxyTypeSSH     ProxyType = "ssh"
	ProxyTypeDirect  ProxyType = "direct"
)

// TunnelResult represents the result of tunnel testing
type TunnelResult struct {
	Target       string            `json:"target"`
	ProxyUsed    string            `json:"proxy_used,omitempty"`
	ProxyType    ProxyType         `json:"proxy_type"`
	Tests        []TunnelTest      `json:"tests"`
	Summary      TunnelSummary     `json:"summary"`
	Duration     time.Duration     `json:"duration"`
}

// TunnelTest represents a single tunnel test
type TunnelTest struct {
	Name         string        `json:"name"`
	Success      bool          `json:"success"`
	Duration     time.Duration `json:"duration"`
	Details      string        `json:"details,omitempty"`
	Error        string        `json:"error,omitempty"`
	ProxyLatency time.Duration `json:"proxy_latency,omitempty"`
	DirectLatency time.Duration `json:"direct_latency,omitempty"`
}

// TunnelSummary contains summary information
type TunnelSummary struct {
	ProxyWorking     bool          `json:"proxy_working"`
	DirectWorking    bool          `json:"direct_working"`
	ProxyFaster      bool          `json:"proxy_faster"`
	LatencyDiff      time.Duration `json:"latency_diff"`
	Recommendation   string        `json:"recommendation"`
}
