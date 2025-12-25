package models

// ConvertFormat represents a file format
type ConvertFormat string

const (
	FormatJSON ConvertFormat = "json"
	FormatYAML ConvertFormat = "yaml"
	FormatTOML ConvertFormat = "toml"
	FormatENV  ConvertFormat = "env"
)

// ConvertResult represents the result of a format conversion
type ConvertResult struct {
	InputFormat  ConvertFormat `json:"input_format"`
	OutputFormat ConvertFormat `json:"output_format"`
	Input        string        `json:"input"`
	Output       string        `json:"output"`
	Error        string        `json:"error,omitempty"`
}
