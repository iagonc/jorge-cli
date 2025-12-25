package models

import "time"

// PortStatus represents the status of a scanned port
type PortStatus string

const (
	PortOpen     PortStatus = "open"
	PortClosed   PortStatus = "closed"
	PortFiltered PortStatus = "filtered"
)

// PortResult contains the result of scanning a single port
type PortResult struct {
	Port    int        `json:"port"`
	Status  PortStatus `json:"status"`
	Service string     `json:"service,omitempty"` // Common service name if known
	Banner  string     `json:"banner,omitempty"`  // Banner grab result if available
}

// PortScanResult contains the complete port scan result
type PortScanResult struct {
	Host          string        `json:"host"`
	StartTime     time.Time     `json:"start_time"`
	EndTime       time.Time     `json:"end_time"`
	Duration      time.Duration `json:"duration"`
	TotalPorts    int           `json:"total_ports"`
	OpenPorts     []PortResult  `json:"open_ports"`
	ClosedPorts   int           `json:"closed_ports"`
	FilteredPorts int           `json:"filtered_ports"`
}

// CommonPorts maps port numbers to common service names
var CommonPorts = map[int]string{
	21:    "FTP",
	22:    "SSH",
	23:    "Telnet",
	25:    "SMTP",
	53:    "DNS",
	80:    "HTTP",
	110:   "POP3",
	143:   "IMAP",
	443:   "HTTPS",
	465:   "SMTPS",
	587:   "SMTP",
	993:   "IMAPS",
	995:   "POP3S",
	3306:  "MySQL",
	3389:  "RDP",
	5432:  "PostgreSQL",
	5672:  "RabbitMQ",
	6379:  "Redis",
	8080:  "HTTP-Alt",
	8443:  "HTTPS-Alt",
	9200:  "Elasticsearch",
	27017: "MongoDB",
}
