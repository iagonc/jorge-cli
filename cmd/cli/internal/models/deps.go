package models

import "time"

// DepsConfig represents the configuration for dependency checks
type DepsConfig struct {
	Name         string       `yaml:"name" json:"name"`
	Dependencies []Dependency `yaml:"dependencies" json:"dependencies"`
}

// Dependency represents a single dependency to check
type Dependency struct {
	Name        string            `yaml:"name" json:"name"`
	Type        DependencyType    `yaml:"type" json:"type"`
	Host        string            `yaml:"host,omitempty" json:"host,omitempty"`
	Port        int               `yaml:"port,omitempty" json:"port,omitempty"`
	URL         string            `yaml:"url,omitempty" json:"url,omitempty"`
	Timeout     int               `yaml:"timeout,omitempty" json:"timeout,omitempty"` // seconds
	Required    bool              `yaml:"required" json:"required"`
	Headers     map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`
	ExpectedStatus int            `yaml:"expected_status,omitempty" json:"expected_status,omitempty"`
}

// DependencyType represents the type of dependency
type DependencyType string

const (
	DepTypeHTTP      DependencyType = "http"
	DepTypeTCP       DependencyType = "tcp"
	DepTypePostgres  DependencyType = "postgres"
	DepTypeMySQL     DependencyType = "mysql"
	DepTypeRedis     DependencyType = "redis"
	DepTypeMongoDB   DependencyType = "mongodb"
	DepTypeGRPC      DependencyType = "grpc"
	DepTypeDNS       DependencyType = "dns"
)

// DepsResult represents the result of dependency checks
type DepsResult struct {
	ConfigName   string             `json:"config_name"`
	TotalDeps    int                `json:"total_deps"`
	HealthyDeps  int                `json:"healthy_deps"`
	UnhealthyDeps int               `json:"unhealthy_deps"`
	Results      []DependencyResult `json:"results"`
	Duration     time.Duration      `json:"duration"`
	AllHealthy   bool               `json:"all_healthy"`
}

// DependencyResult represents the result of checking a single dependency
type DependencyResult struct {
	Name         string        `json:"name"`
	Type         DependencyType `json:"type"`
	Healthy      bool          `json:"healthy"`
	ResponseTime time.Duration `json:"response_time"`
	Details      string        `json:"details,omitempty"`
	Error        string        `json:"error,omitempty"`
	Required     bool          `json:"required"`
}
