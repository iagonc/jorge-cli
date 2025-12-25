package models

// SecretScanResult represents the result of a secret scan
type SecretScanResult struct {
	Path          string         `json:"path"`
	TotalFiles    int            `json:"total_files"`
	ScannedFiles  int            `json:"scanned_files"`
	SkippedFiles  int            `json:"skipped_files"`
	Findings      []SecretFinding `json:"findings"`
	TotalFindings int            `json:"total_findings"`
}

// SecretFinding represents a found secret
type SecretFinding struct {
	File        string     `json:"file"`
	Line        int        `json:"line"`
	Type        SecretType `json:"type"`
	Description string     `json:"description"`
	Match       string     `json:"match"` // Redacted match
	Severity    string     `json:"severity"` // high, medium, low
}

// SecretType represents the type of secret
type SecretType string

const (
	SecretTypeAWSKey        SecretType = "aws_key"
	SecretTypeAWSSecret     SecretType = "aws_secret"
	SecretTypeGCPKey        SecretType = "gcp_key"
	SecretTypeGitHubToken   SecretType = "github_token"
	SecretTypeSlackToken    SecretType = "slack_token"
	SecretTypePrivateKey    SecretType = "private_key"
	SecretTypePassword      SecretType = "password"
	SecretTypeAPIKey        SecretType = "api_key"
	SecretTypeJWT           SecretType = "jwt"
	SecretTypeGeneric       SecretType = "generic_secret"
)

// SecretPattern represents a pattern to match secrets
type SecretPattern struct {
	Type        SecretType
	Pattern     string
	Description string
	Severity    string
}
