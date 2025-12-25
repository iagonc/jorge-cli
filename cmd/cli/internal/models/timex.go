package models

import "time"

// TimeConversion represents a time conversion result
type TimeConversion struct {
	Input       string            `json:"input"`
	InputFormat string            `json:"input_format"`
	Parsed      time.Time         `json:"parsed"`
	Outputs     map[string]string `json:"outputs"`
	Timezones   []TimezoneResult  `json:"timezones,omitempty"`
}

// TimezoneResult represents time in a specific timezone
type TimezoneResult struct {
	Timezone string `json:"timezone"`
	Time     string `json:"time"`
	Offset   string `json:"offset"`
}

// TimeNow represents the current time in various formats
type TimeNow struct {
	Unix      int64             `json:"unix"`
	UnixMilli int64             `json:"unix_milli"`
	UnixNano  int64             `json:"unix_nano"`
	ISO8601   string            `json:"iso8601"`
	RFC3339   string            `json:"rfc3339"`
	RFC1123   string            `json:"rfc1123"`
	Human     string            `json:"human"`
	Timezones []TimezoneResult  `json:"timezones,omitempty"`
}
