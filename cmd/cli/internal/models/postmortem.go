package models

import "time"

// Postmortem represents a post-incident review document
type Postmortem struct {
	ID              string          `json:"id" yaml:"id"`
	IncidentID      string          `json:"incident_id" yaml:"incident_id"`
	Title           string          `json:"title" yaml:"title"`
	Date            time.Time       `json:"date" yaml:"date"`
	Authors         []string        `json:"authors" yaml:"authors"`
	Status          PostmortemStatus `json:"status" yaml:"status"`
	Severity        string          `json:"severity" yaml:"severity"`

	// Summary
	Summary         string          `json:"summary" yaml:"summary"`
	Impact          ImpactSummary   `json:"impact" yaml:"impact"`

	// Timeline
	Timeline        []TimelineEntry `json:"timeline" yaml:"timeline"`

	// Root Cause Analysis
	RootCauses      []string        `json:"root_causes" yaml:"root_causes"`
	Contributing    []string        `json:"contributing_factors,omitempty" yaml:"contributing_factors,omitempty"`

	// What went well/poorly
	WentWell        []string        `json:"went_well,omitempty" yaml:"went_well,omitempty"`
	WentPoorly      []string        `json:"went_poorly,omitempty" yaml:"went_poorly,omitempty"`
	LuckyFactors    []string        `json:"lucky_factors,omitempty" yaml:"lucky_factors,omitempty"`

	// Action Items
	ActionItems     []ActionItem    `json:"action_items" yaml:"action_items"`

	// Lessons Learned
	LessonsLearned  []string        `json:"lessons_learned,omitempty" yaml:"lessons_learned,omitempty"`

	// Metadata
	CreatedAt       time.Time       `json:"created_at" yaml:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" yaml:"updated_at"`
	ReviewedAt      *time.Time      `json:"reviewed_at,omitempty" yaml:"reviewed_at,omitempty"`
}

// PostmortemStatus represents the status of a postmortem
type PostmortemStatus string

const (
	PostmortemDraft     PostmortemStatus = "draft"
	PostmortemInReview  PostmortemStatus = "in_review"
	PostmortemApproved  PostmortemStatus = "approved"
	PostmortemPublished PostmortemStatus = "published"
)

// ImpactSummary represents the impact of an incident
type ImpactSummary struct {
	Duration        string   `json:"duration" yaml:"duration"`
	UsersAffected   string   `json:"users_affected,omitempty" yaml:"users_affected,omitempty"`
	RequestsAffected string  `json:"requests_affected,omitempty" yaml:"requests_affected,omitempty"`
	Revenue         string   `json:"revenue,omitempty" yaml:"revenue,omitempty"`
	SLABreach       bool     `json:"sla_breach" yaml:"sla_breach"`
	ServicesAffected []string `json:"services_affected" yaml:"services_affected"`
}

// TimelineEntry represents an entry in the incident timeline
type TimelineEntry struct {
	Time        time.Time `json:"time" yaml:"time"`
	Description string    `json:"description" yaml:"description"`
	Type        string    `json:"type,omitempty" yaml:"type,omitempty"` // detection, response, mitigation, resolution
}

// ActionItem represents a follow-up action from the postmortem
type ActionItem struct {
	ID          string     `json:"id" yaml:"id"`
	Description string     `json:"description" yaml:"description"`
	Type        string     `json:"type" yaml:"type"` // prevent, detect, mitigate, process
	Priority    string     `json:"priority" yaml:"priority"` // P0, P1, P2, P3
	Owner       string     `json:"owner" yaml:"owner"`
	DueDate     *time.Time `json:"due_date,omitempty" yaml:"due_date,omitempty"`
	Status      string     `json:"status" yaml:"status"` // todo, in_progress, done
	TicketURL   string     `json:"ticket_url,omitempty" yaml:"ticket_url,omitempty"`
}

// PostmortemTemplate represents a template for creating postmortems
type PostmortemTemplate struct {
	Name        string   `json:"name" yaml:"name"`
	Sections    []string `json:"sections" yaml:"sections"`
	Questions   []string `json:"questions,omitempty" yaml:"questions,omitempty"`
}

// PostmortemList represents a list of postmortems
type PostmortemList struct {
	Postmortems []Postmortem `json:"postmortems"`
	Total       int          `json:"total"`
}
