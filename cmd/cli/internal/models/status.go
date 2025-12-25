package models

import "time"

// StatusPageConfig represents status page configuration
type StatusPageConfig struct {
	Title       string            `json:"title" yaml:"title"`
	Description string            `json:"description" yaml:"description"`
	Components  []StatusComponent `json:"components" yaml:"components"`
	Incidents   []StatusIncident  `json:"incidents,omitempty" yaml:"incidents,omitempty"`
}

// StatusComponent represents a component to monitor
type StatusComponent struct {
	Name        string       `json:"name" yaml:"name"`
	Description string       `json:"description,omitempty" yaml:"description,omitempty"`
	Type        string       `json:"type" yaml:"type"` // http, tcp, dns, custom
	Target      string       `json:"target" yaml:"target"`
	Status      ComponentStatus `json:"status,omitempty"`
	Group       string       `json:"group,omitempty" yaml:"group,omitempty"`
}

// ComponentStatus represents the current status of a component
type ComponentStatus struct {
	State       StatusState `json:"state"`
	Latency     string      `json:"latency,omitempty"`
	LastChecked time.Time   `json:"last_checked"`
	Message     string      `json:"message,omitempty"`
}

// StatusState represents the state of a component
type StatusState string

const (
	StatusOperational     StatusState = "operational"
	StatusDegraded        StatusState = "degraded"
	StatusPartialOutage   StatusState = "partial_outage"
	StatusMajorOutage     StatusState = "major_outage"
	StatusMaintenance     StatusState = "maintenance"
	StatusUnknown         StatusState = "unknown"
)

// StatusIncident represents an incident on the status page
type StatusIncident struct {
	ID          string      `json:"id" yaml:"id"`
	Title       string      `json:"title" yaml:"title"`
	Status      string      `json:"status" yaml:"status"` // investigating, identified, monitoring, resolved
	Impact      string      `json:"impact" yaml:"impact"` // none, minor, major, critical
	CreatedAt   time.Time   `json:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" yaml:"updated_at"`
	ResolvedAt  *time.Time  `json:"resolved_at,omitempty" yaml:"resolved_at,omitempty"`
	Updates     []IncidentUpdate `json:"updates,omitempty" yaml:"updates,omitempty"`
	Components  []string    `json:"components,omitempty" yaml:"components,omitempty"`
}

// IncidentUpdate represents an update to an incident
type IncidentUpdate struct {
	Status    string    `json:"status" yaml:"status"`
	Message   string    `json:"message" yaml:"message"`
	CreatedAt time.Time `json:"created_at" yaml:"created_at"`
}

// StatusPageResult represents the overall status page result
type StatusPageResult struct {
	Title         string            `json:"title"`
	Description   string            `json:"description"`
	OverallStatus StatusState       `json:"overall_status"`
	Components    []StatusComponent `json:"components"`
	ActiveIncidents []StatusIncident `json:"active_incidents,omitempty"`
	LastUpdated   time.Time         `json:"last_updated"`
}
