package models

// DiffResult represents the result of a diff operation
type DiffResult struct {
	File1     string       `json:"file1"`
	File2     string       `json:"file2"`
	Identical bool         `json:"identical"`
	Changes   []DiffChange `json:"changes"`
	Stats     DiffStats    `json:"stats"`
}

// DiffChange represents a single change
type DiffChange struct {
	Type     DiffType `json:"type"`
	Path     string   `json:"path,omitempty"` // For structured diffs
	Line     int      `json:"line,omitempty"` // For text diffs
	OldValue string   `json:"old_value,omitempty"`
	NewValue string   `json:"new_value,omitempty"`
}

// DiffType represents the type of change
type DiffType string

const (
	DiffTypeAdded    DiffType = "added"
	DiffTypeRemoved  DiffType = "removed"
	DiffTypeModified DiffType = "modified"
)

// DiffStats represents diff statistics
type DiffStats struct {
	Added    int `json:"added"`
	Removed  int `json:"removed"`
	Modified int `json:"modified"`
}
