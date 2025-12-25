package models

import "time"

// Incident represents an incident record
type Incident struct {
	ID          string          `json:"id" yaml:"id"`
	Title       string          `json:"title" yaml:"title"`
	Description string          `json:"description" yaml:"description"`
	Severity    IncidentSeverity `json:"severity" yaml:"severity"`
	Status      IncidentStatus  `json:"status" yaml:"status"`
	Commander   string          `json:"commander,omitempty" yaml:"commander,omitempty"`
	Team        []string        `json:"team,omitempty" yaml:"team,omitempty"`
	Services    []string        `json:"services,omitempty" yaml:"services,omitempty"`
	Timeline    []TimelineEvent `json:"timeline,omitempty" yaml:"timeline,omitempty"`
	CreatedAt   time.Time       `json:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" yaml:"updated_at"`
	ResolvedAt  *time.Time      `json:"resolved_at,omitempty" yaml:"resolved_at,omitempty"`
	Duration    string          `json:"duration,omitempty" yaml:"duration,omitempty"`
	Tags        []string        `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// IncidentSeverity represents incident severity level
type IncidentSeverity string

const (
	IncidentSeverityCritical IncidentSeverity = "critical" // SEV1
	IncidentSeverityHigh     IncidentSeverity = "high"     // SEV2
	IncidentSeverityMedium   IncidentSeverity = "medium"   // SEV3
	IncidentSeverityLow      IncidentSeverity = "low"      // SEV4
)

// IncidentStatus represents incident status
type IncidentStatus string

const (
	StatusOpen         IncidentStatus = "open"
	StatusInvestigating IncidentStatus = "investigating"
	StatusIdentified   IncidentStatus = "identified"
	StatusMonitoring   IncidentStatus = "monitoring"
	StatusResolved     IncidentStatus = "resolved"
)

// TimelineEvent represents an event in the incident timeline
type TimelineEvent struct {
	Timestamp   time.Time `json:"timestamp" yaml:"timestamp"`
	Type        string    `json:"type" yaml:"type"` // status_change, note, action, alert
	Description string    `json:"description" yaml:"description"`
	Author      string    `json:"author,omitempty" yaml:"author,omitempty"`
}

// IncidentList represents a list of incidents
type IncidentList struct {
	Incidents []Incident `json:"incidents"`
	Total     int        `json:"total"`
	Open      int        `json:"open"`
	Resolved  int        `json:"resolved"`
}

// IncidentSummary represents incident summary stats
type IncidentSummary struct {
	TotalIncidents int            `json:"total_incidents"`
	ByStatus       map[string]int `json:"by_status"`
	BySeverity     map[string]int `json:"by_severity"`
	MTTR           string         `json:"mttr"` // Mean Time To Resolution
	AvgDuration    string         `json:"avg_duration"`
}
