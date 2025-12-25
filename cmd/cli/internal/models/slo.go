package models

import "time"

// SLO represents a Service Level Objective
type SLO struct {
	Name        string       `json:"name" yaml:"name"`
	Description string       `json:"description,omitempty" yaml:"description,omitempty"`
	Service     string       `json:"service" yaml:"service"`
	SLI         SLI          `json:"sli" yaml:"sli"`
	Target      float64      `json:"target" yaml:"target"` // e.g., 99.9
	Window      string       `json:"window" yaml:"window"` // e.g., "30d", "7d"
	Budgets     []ErrorBudget `json:"budgets,omitempty" yaml:"budgets,omitempty"`
}

// SLI represents a Service Level Indicator
type SLI struct {
	Type       string `json:"type" yaml:"type"` // availability, latency, error_rate, throughput
	Query      string `json:"query,omitempty" yaml:"query,omitempty"`
	GoodEvents string `json:"good_events,omitempty" yaml:"good_events,omitempty"`
	TotalEvents string `json:"total_events,omitempty" yaml:"total_events,omitempty"`
	Threshold  string `json:"threshold,omitempty" yaml:"threshold,omitempty"` // e.g., "200ms" for latency
}

// ErrorBudget represents the error budget calculation
type ErrorBudget struct {
	Window        string    `json:"window"`
	Target        float64   `json:"target"`
	Current       float64   `json:"current"`
	BudgetTotal   float64   `json:"budget_total"`   // Total allowed downtime/errors
	BudgetUsed    float64   `json:"budget_used"`    // Used budget
	BudgetRemaining float64 `json:"budget_remaining"`
	BudgetPercent float64   `json:"budget_percent"` // Percentage of budget remaining
	Status        string    `json:"status"`         // healthy, warning, critical, exhausted
	CalculatedAt  time.Time `json:"calculated_at"`
}

// SLOReport represents an SLO compliance report
type SLOReport struct {
	SLO           SLO         `json:"slo"`
	Period        string      `json:"period"`
	StartTime     time.Time   `json:"start_time"`
	EndTime       time.Time   `json:"end_time"`
	TotalEvents   int64       `json:"total_events"`
	GoodEvents    int64       `json:"good_events"`
	BadEvents     int64       `json:"bad_events"`
	CurrentSLI    float64     `json:"current_sli"`
	Target        float64     `json:"target"`
	Compliance    bool        `json:"compliance"`
	ErrorBudget   ErrorBudget `json:"error_budget"`
	Trend         string      `json:"trend"` // improving, stable, degrading
	Incidents     int         `json:"incidents"`
}

// SLOConfig represents SLO configuration file
type SLOConfig struct {
	Version  string `json:"version" yaml:"version"`
	Service  string `json:"service" yaml:"service"`
	SLOs     []SLO  `json:"slos" yaml:"slos"`
}

// SLODashboard represents a dashboard view of multiple SLOs
type SLODashboard struct {
	Service     string      `json:"service"`
	SLOs        []SLOReport `json:"slos"`
	OverallHealth string    `json:"overall_health"` // healthy, warning, critical
	LastUpdated time.Time   `json:"last_updated"`
}
