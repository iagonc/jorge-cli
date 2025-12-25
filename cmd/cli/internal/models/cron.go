package models

import "time"

// CronExpression represents a parsed cron expression
type CronExpression struct {
	Expression  string      `json:"expression"`
	Description string      `json:"description"`
	NextRuns    []time.Time `json:"next_runs,omitempty"`
	IsValid     bool        `json:"is_valid"`
	Error       string      `json:"error,omitempty"`
	Fields      *CronFields `json:"fields,omitempty"`
}

// CronFields represents the parsed fields of a cron expression
type CronFields struct {
	Minute     string `json:"minute"`
	Hour       string `json:"hour"`
	DayOfMonth string `json:"day_of_month"`
	Month      string `json:"month"`
	DayOfWeek  string `json:"day_of_week"`
}
