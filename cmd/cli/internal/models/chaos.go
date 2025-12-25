package models

import "time"

// ChaosExperiment represents a chaos engineering experiment
type ChaosExperiment struct {
	ID          string           `json:"id" yaml:"id"`
	Name        string           `json:"name" yaml:"name"`
	Description string           `json:"description,omitempty" yaml:"description,omitempty"`
	Type        ChaosType        `json:"type" yaml:"type"`
	Target      ChaosTarget      `json:"target" yaml:"target"`
	Config      ChaosConfig      `json:"config" yaml:"config"`
	Duration    string           `json:"duration" yaml:"duration"`
	Status      ExperimentStatus `json:"status"`
	StartedAt   *time.Time       `json:"started_at,omitempty"`
	EndedAt     *time.Time       `json:"ended_at,omitempty"`
	Results     *ChaosResults    `json:"results,omitempty"`
}

// ChaosType represents the type of chaos experiment
type ChaosType string

const (
	ChaosTypeNetworkLatency  ChaosType = "network_latency"
	ChaosTypeNetworkLoss     ChaosType = "network_loss"
	ChaosTypeNetworkPartition ChaosType = "network_partition"
	ChaosTypeCPUStress       ChaosType = "cpu_stress"
	ChaosTypeMemoryStress    ChaosType = "memory_stress"
	ChaosTypeDiskStress      ChaosType = "disk_stress"
	ChaosTypeProcessKill     ChaosType = "process_kill"
	ChaosTypeHTTPFault       ChaosType = "http_fault"
	ChaosTypeDNSFault        ChaosType = "dns_fault"
)

// ExperimentStatus represents the status of an experiment
type ExperimentStatus string

const (
	ExperimentPending   ExperimentStatus = "pending"
	ExperimentRunning   ExperimentStatus = "running"
	ExperimentCompleted ExperimentStatus = "completed"
	ExperimentFailed    ExperimentStatus = "failed"
	ExperimentAborted   ExperimentStatus = "aborted"
)

// ChaosTarget represents the target of a chaos experiment
type ChaosTarget struct {
	Type      string            `json:"type" yaml:"type"` // host, container, pod, service
	Selector  map[string]string `json:"selector,omitempty" yaml:"selector,omitempty"`
	Hosts     []string          `json:"hosts,omitempty" yaml:"hosts,omitempty"`
	Namespace string            `json:"namespace,omitempty" yaml:"namespace,omitempty"`
}

// ChaosConfig represents experiment-specific configuration
type ChaosConfig struct {
	// Network
	Latency     string `json:"latency,omitempty" yaml:"latency,omitempty"`         // e.g., "100ms"
	Jitter      string `json:"jitter,omitempty" yaml:"jitter,omitempty"`           // e.g., "20ms"
	PacketLoss  int    `json:"packet_loss,omitempty" yaml:"packet_loss,omitempty"` // percentage
	Bandwidth   string `json:"bandwidth,omitempty" yaml:"bandwidth,omitempty"`     // e.g., "1mbps"
	Port        int    `json:"port,omitempty" yaml:"port,omitempty"`
	TargetHosts []string `json:"target_hosts,omitempty" yaml:"target_hosts,omitempty"`

	// Resource stress
	CPUPercent    int    `json:"cpu_percent,omitempty" yaml:"cpu_percent,omitempty"`
	MemoryBytes   string `json:"memory_bytes,omitempty" yaml:"memory_bytes,omitempty"`
	DiskFillBytes string `json:"disk_fill_bytes,omitempty" yaml:"disk_fill_bytes,omitempty"`
	Workers       int    `json:"workers,omitempty" yaml:"workers,omitempty"`

	// Process
	ProcessName string `json:"process_name,omitempty" yaml:"process_name,omitempty"`
	Signal      string `json:"signal,omitempty" yaml:"signal,omitempty"` // SIGTERM, SIGKILL

	// HTTP
	StatusCode int    `json:"status_code,omitempty" yaml:"status_code,omitempty"`
	ErrorRate  int    `json:"error_rate,omitempty" yaml:"error_rate,omitempty"` // percentage
	Delay      string `json:"delay,omitempty" yaml:"delay,omitempty"`
}

// ChaosResults represents the results of a chaos experiment
type ChaosResults struct {
	Success     bool              `json:"success"`
	Summary     string            `json:"summary"`
	Metrics     map[string]string `json:"metrics,omitempty"`
	Observations []string         `json:"observations,omitempty"`
	Errors      []string          `json:"errors,omitempty"`
}

// ChaosScenario represents a predefined chaos scenario
type ChaosScenario struct {
	Name        string            `json:"name" yaml:"name"`
	Description string            `json:"description" yaml:"description"`
	Experiments []ChaosExperiment `json:"experiments" yaml:"experiments"`
	Steady      []SteadyStateCheck `json:"steady_state,omitempty" yaml:"steady_state,omitempty"`
}

// SteadyStateCheck represents a steady state verification
type SteadyStateCheck struct {
	Name      string `json:"name" yaml:"name"`
	Type      string `json:"type" yaml:"type"` // http, tcp, command
	Target    string `json:"target" yaml:"target"`
	Expect    string `json:"expect" yaml:"expect"`
	Tolerance string `json:"tolerance,omitempty" yaml:"tolerance,omitempty"`
}
