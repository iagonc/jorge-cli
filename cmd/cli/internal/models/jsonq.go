package models

// JSONQuery represents a JSON query operation
type JSONQuery struct {
	Input   string      `json:"input"`
	Query   string      `json:"query"`
	Result  interface{} `json:"result"`
	Error   string      `json:"error,omitempty"`
	Type    string      `json:"type"` // string, number, boolean, object, array, null
}

// JSONDiff represents differences between two JSON documents
type JSONDiff struct {
	Path     string      `json:"path"`
	Type     string      `json:"type"` // added, removed, changed
	OldValue interface{} `json:"old_value,omitempty"`
	NewValue interface{} `json:"new_value,omitempty"`
}
