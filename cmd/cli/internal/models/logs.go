package models

import "time"

// LogEntry represents a parsed log entry
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp,omitempty"`
	Level     string                 `json:"level,omitempty"`
	Message   string                 `json:"message"`
	Source    string                 `json:"source,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Raw       string                 `json:"raw"`
	LineNum   int                    `json:"line_num"`
}

// LogLevel represents log severity
type LogLevel string

const (
	LogLevelDebug   LogLevel = "DEBUG"
	LogLevelInfo    LogLevel = "INFO"
	LogLevelWarn    LogLevel = "WARN"
	LogLevelError   LogLevel = "ERROR"
	LogLevelFatal   LogLevel = "FATAL"
	LogLevelUnknown LogLevel = "UNKNOWN"
)

// LogStats represents log statistics
type LogStats struct {
	TotalLines   int                `json:"total_lines"`
	ParsedLines  int                `json:"parsed_lines"`
	FailedLines  int                `json:"failed_lines"`
	LevelCounts  map[string]int     `json:"level_counts"`
	ErrorRate    float64            `json:"error_rate"`
	TimeRange    *TimeRange         `json:"time_range,omitempty"`
	TopErrors    []ErrorCount       `json:"top_errors,omitempty"`
	TopSources   []SourceCount      `json:"top_sources,omitempty"`
}

// TimeRange represents a time range
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// ErrorCount represents error frequency
type ErrorCount struct {
	Message string `json:"message"`
	Count   int    `json:"count"`
}

// SourceCount represents source frequency
type SourceCount struct {
	Source string `json:"source"`
	Count  int    `json:"count"`
}

// LogFormat represents the log format
type LogFormat string

const (
	LogFormatJSON    LogFormat = "json"
	LogFormatNginx   LogFormat = "nginx"
	LogFormatApache  LogFormat = "apache"
	LogFormatSyslog  LogFormat = "syslog"
	LogFormatCommon  LogFormat = "common"
	LogFormatUnknown LogFormat = "unknown"
)
