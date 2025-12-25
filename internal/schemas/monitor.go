package schemas

import (
	"time"

	"gorm.io/gorm"
)

// MonitorType defines the type of monitoring check
type MonitorType string

const (
	MonitorTypeHTTP  MonitorType = "http"
	MonitorTypeTCP   MonitorType = "tcp"
	MonitorTypePing  MonitorType = "ping"
	MonitorTypeDNS   MonitorType = "dns"
	MonitorTypeSSL   MonitorType = "ssl"
	MonitorTypeTrace MonitorType = "trace"
)

// MonitorStatus represents the current status of a monitor
type MonitorStatus string

const (
	MonitorStatusUp       MonitorStatus = "up"
	MonitorStatusDown     MonitorStatus = "down"
	MonitorStatusDegraded MonitorStatus = "degraded"
	MonitorStatusUnknown  MonitorStatus = "unknown"
)

// Monitor represents an endpoint being monitored
type Monitor struct {
	gorm.Model
	Name        string        `json:"name" gorm:"not null"`
	Target      string        `json:"target" gorm:"not null"` // URL, IP, or hostname
	Type        MonitorType   `json:"type" gorm:"not null;default:'http'"`
	Port        int           `json:"port" gorm:"default:0"`                  // For TCP checks
	Interval    int           `json:"interval" gorm:"not null;default:30"`    // Check interval in seconds
	Timeout     int           `json:"timeout" gorm:"not null;default:10"`     // Timeout in seconds
	Enabled     bool          `json:"enabled" gorm:"not null;default:true"`
	Status      MonitorStatus `json:"status" gorm:"not null;default:'unknown'"`
	LastCheckAt *time.Time    `json:"last_check_at"`
	LastLatency int64         `json:"last_latency"` // Latency in milliseconds
	Uptime      float64       `json:"uptime" gorm:"default:0"` // Uptime percentage

	// HTTP specific
	ExpectedStatus int    `json:"expected_status" gorm:"default:200"`
	Method         string `json:"method" gorm:"default:'GET'"`

	// Relations
	Results []CheckResult `json:"results,omitempty" gorm:"foreignKey:MonitorID"`
	Alerts  []Alert       `json:"alerts,omitempty" gorm:"foreignKey:MonitorID"`
}

// CheckResult stores the result of a monitoring check
type CheckResult struct {
	gorm.Model
	MonitorID    uint          `json:"monitor_id" gorm:"not null;index"`
	Status       MonitorStatus `json:"status" gorm:"not null"`
	Latency      int64         `json:"latency"` // in milliseconds
	StatusCode   int           `json:"status_code,omitempty"` // HTTP status code
	Output       string        `json:"output" gorm:"type:text"` // Raw output from the check
	ErrorMessage string        `json:"error_message,omitempty"`
	CheckedAt    time.Time     `json:"checked_at" gorm:"not null;index"`

	// DNS specific
	ResolvedIPs string `json:"resolved_ips,omitempty"`

	// SSL specific
	CertExpiry    *time.Time `json:"cert_expiry,omitempty"`
	CertIssuer    string     `json:"cert_issuer,omitempty"`
	CertDaysLeft  int        `json:"cert_days_left,omitempty"`
}

// AlertSeverity defines the severity level of an alert
type AlertSeverity string

const (
	AlertSeverityCritical AlertSeverity = "critical"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityInfo     AlertSeverity = "info"
)

// Alert represents an alert triggered by monitoring
type Alert struct {
	gorm.Model
	MonitorID    uint          `json:"monitor_id" gorm:"not null;index"`
	Severity     AlertSeverity `json:"severity" gorm:"not null"`
	Title        string        `json:"title" gorm:"not null"`
	Message      string        `json:"message" gorm:"type:text"`
	Acknowledged bool          `json:"acknowledged" gorm:"default:false"`
	AckedAt      *time.Time    `json:"acked_at,omitempty"`
	ResolvedAt   *time.Time    `json:"resolved_at,omitempty"`

	// Relations
	Monitor Monitor `json:"monitor,omitempty" gorm:"foreignKey:MonitorID"`
}

// DiagnosticRequest represents a one-time diagnostic request
type DiagnosticRequest struct {
	Target  string `json:"target" binding:"required"`
	Type    string `json:"type" binding:"required"` // ping, dns, tcp, http, trace, ssl
	Port    int    `json:"port,omitempty"`
	Timeout int    `json:"timeout,omitempty"`
}

// DiagnosticResult represents the result of a diagnostic
type DiagnosticResult struct {
	Success  bool     `json:"success"`
	Type     string   `json:"type"`
	Target   string   `json:"target"`
	Latency  int64    `json:"latency_ms"`
	Output   string   `json:"output"`
	Error    string   `json:"error,omitempty"`
	Details  map[string]interface{} `json:"details,omitempty"`
}
