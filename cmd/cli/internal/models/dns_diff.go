package models

import "time"

// DNSRecordType represents the type of DNS record
type DNSRecordType string

const (
	RecordTypeA     DNSRecordType = "A"
	RecordTypeAAAA  DNSRecordType = "AAAA"
	RecordTypeCNAME DNSRecordType = "CNAME"
	RecordTypeMX    DNSRecordType = "MX"
	RecordTypeTXT   DNSRecordType = "TXT"
	RecordTypeNS    DNSRecordType = "NS"
)

// ExpectedRecord represents an expected DNS record in the config
type ExpectedRecord struct {
	Domain     string        `yaml:"domain" json:"domain"`
	RecordType DNSRecordType `yaml:"type" json:"type"`
	Values     []string      `yaml:"values" json:"values"`
}

// DNSDiffConfig represents the YAML configuration for DNS diff
type DNSDiffConfig struct {
	Nameserver string           `yaml:"nameserver,omitempty" json:"nameserver,omitempty"`
	Records    []ExpectedRecord `yaml:"records" json:"records"`
}

// RecordDiff represents the difference between expected and actual records
type RecordDiff struct {
	Domain     string        `json:"domain"`
	RecordType DNSRecordType `json:"record_type"`
	Expected   []string      `json:"expected"`
	Actual     []string      `json:"actual"`
	Match      bool          `json:"match"`
	Missing    []string      `json:"missing,omitempty"` // In expected but not actual
	Extra      []string      `json:"extra,omitempty"`   // In actual but not expected
	Error      string        `json:"error,omitempty"`
}

// DNSDiffResult contains the complete DNS diff result
type DNSDiffResult struct {
	Nameserver   string        `json:"nameserver"`
	TotalRecords int           `json:"total_records"`
	Matching     int           `json:"matching"`
	Mismatched   int           `json:"mismatched"`
	Errors       int           `json:"errors"`
	Diffs        []RecordDiff  `json:"diffs"`
	Duration     time.Duration `json:"duration"`
}
