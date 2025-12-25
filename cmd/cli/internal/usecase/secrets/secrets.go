package secrets

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"go.uber.org/zap"
)

// SecretsUsecase handles secret scanning
type SecretsUsecase struct {
	logger   *zap.Logger
	patterns []models.SecretPattern
}

// NewSecretsUsecase creates a new SecretsUsecase
func NewSecretsUsecase(logger *zap.Logger) *SecretsUsecase {
	return &SecretsUsecase{
		logger:   logger,
		patterns: defaultPatterns(),
	}
}

func defaultPatterns() []models.SecretPattern {
	return []models.SecretPattern{
		// AWS
		{Type: models.SecretTypeAWSKey, Pattern: `AKIA[0-9A-Z]{16}`, Description: "AWS Access Key ID", Severity: "high"},
		{Type: models.SecretTypeAWSSecret, Pattern: `(?i)aws_secret_access_key\s*[=:]\s*['"]?([A-Za-z0-9/+=]{40})['"]?`, Description: "AWS Secret Access Key", Severity: "high"},

		// GitHub
		{Type: models.SecretTypeGitHubToken, Pattern: `ghp_[a-zA-Z0-9]{36}`, Description: "GitHub Personal Access Token", Severity: "high"},
		{Type: models.SecretTypeGitHubToken, Pattern: `github_pat_[a-zA-Z0-9]{22}_[a-zA-Z0-9]{59}`, Description: "GitHub Fine-grained Token", Severity: "high"},

		// Slack
		{Type: models.SecretTypeSlackToken, Pattern: `xox[baprs]-[0-9]{10,13}-[0-9]{10,13}[a-zA-Z0-9-]*`, Description: "Slack Token", Severity: "high"},

		// Private Keys
		{Type: models.SecretTypePrivateKey, Pattern: `-----BEGIN (RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----`, Description: "Private Key", Severity: "high"},
		{Type: models.SecretTypePrivateKey, Pattern: `-----BEGIN PGP PRIVATE KEY BLOCK-----`, Description: "PGP Private Key", Severity: "high"},

		// API Keys
		{Type: models.SecretTypeAPIKey, Pattern: `(?i)(api[_-]?key|apikey)\s*[=:]\s*['"]?([a-zA-Z0-9]{20,})['"]?`, Description: "API Key", Severity: "medium"},
		{Type: models.SecretTypeAPIKey, Pattern: `(?i)(api[_-]?secret|apisecret)\s*[=:]\s*['"]?([a-zA-Z0-9]{20,})['"]?`, Description: "API Secret", Severity: "high"},

		// JWT
		{Type: models.SecretTypeJWT, Pattern: `eyJ[a-zA-Z0-9]{10,}\.eyJ[a-zA-Z0-9]{10,}\.[a-zA-Z0-9_-]{10,}`, Description: "JWT Token", Severity: "medium"},

		// Passwords
		{Type: models.SecretTypePassword, Pattern: `(?i)(password|passwd|pwd)\s*[=:]\s*['"]?([^\s'"]{8,})['"]?`, Description: "Password", Severity: "high"},
		{Type: models.SecretTypePassword, Pattern: `(?i)(db_password|database_password|mysql_password|postgres_password)\s*[=:]\s*['"]?([^\s'"]+)['"]?`, Description: "Database Password", Severity: "high"},

		// Generic Secrets
		{Type: models.SecretTypeGeneric, Pattern: `(?i)(secret|token|auth)\s*[=:]\s*['"]?([a-zA-Z0-9/+=]{20,})['"]?`, Description: "Generic Secret", Severity: "medium"},

		// GCP
		{Type: models.SecretTypeGCPKey, Pattern: `"type":\s*"service_account"`, Description: "GCP Service Account Key", Severity: "high"},
	}
}

// Scan scans a path for secrets
func (u *SecretsUsecase) Scan(path string, excludes []string) (*models.SecretScanResult, error) {
	result := &models.SecretScanResult{
		Path:     path,
		Findings: []models.SecretFinding{},
	}

	// Compile patterns
	compiledPatterns := make([]*regexp.Regexp, 0, len(u.patterns))
	for _, p := range u.patterns {
		if re, err := regexp.Compile(p.Pattern); err == nil {
			compiledPatterns = append(compiledPatterns, re)
		}
	}

	// Walk directory
	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip directories
		if info.IsDir() {
			// Skip common directories
			base := filepath.Base(filePath)
			if base == ".git" || base == "node_modules" || base == "vendor" || base == ".venv" || base == "__pycache__" {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip large files
		if info.Size() > 10*1024*1024 { // 10MB
			result.SkippedFiles++
			return nil
		}

		// Skip binary files
		ext := strings.ToLower(filepath.Ext(filePath))
		binaryExts := []string{".exe", ".dll", ".so", ".dylib", ".bin", ".png", ".jpg", ".jpeg", ".gif", ".ico", ".pdf", ".zip", ".tar", ".gz", ".rar"}
		for _, be := range binaryExts {
			if ext == be {
				result.SkippedFiles++
				return nil
			}
		}

		// Check excludes
		for _, exclude := range excludes {
			if matched, _ := filepath.Match(exclude, filepath.Base(filePath)); matched {
				result.SkippedFiles++
				return nil
			}
		}

		result.TotalFiles++
		findings := u.scanFile(filePath, compiledPatterns)
		result.Findings = append(result.Findings, findings...)
		result.ScannedFiles++

		return nil
	})

	result.TotalFindings = len(result.Findings)
	return result, err
}

func (u *SecretsUsecase) scanFile(filePath string, patterns []*regexp.Regexp) []models.SecretFinding {
	var findings []models.SecretFinding

	file, err := os.Open(filePath)
	if err != nil {
		return findings
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		for i, pattern := range patterns {
			if matches := pattern.FindStringSubmatch(line); matches != nil {
				// Redact the match
				match := matches[0]
				if len(match) > 20 {
					match = match[:10] + "..." + match[len(match)-5:]
				}

				findings = append(findings, models.SecretFinding{
					File:        filePath,
					Line:        lineNum,
					Type:        u.patterns[i].Type,
					Description: u.patterns[i].Description,
					Match:       match,
					Severity:    u.patterns[i].Severity,
				})
				break // One finding per line
			}
		}
	}

	return findings
}

// ScanFile scans a single file
func (u *SecretsUsecase) ScanFile(filePath string) ([]models.SecretFinding, error) {
	compiledPatterns := make([]*regexp.Regexp, 0, len(u.patterns))
	for _, p := range u.patterns {
		if re, err := regexp.Compile(p.Pattern); err == nil {
			compiledPatterns = append(compiledPatterns, re)
		}
	}

	return u.scanFile(filePath, compiledPatterns), nil
}

// GetPatterns returns all secret patterns
func (u *SecretsUsecase) GetPatterns() []models.SecretPattern {
	return u.patterns
}
