package playbook

import (
	"fmt"
	"regexp"
	"strings"
)

// SecurityValidator validates commands for safety
type SecurityValidator struct {
	blockedCommands []string
	blockedPatterns []*regexp.Regexp
	dangerousPatterns []string
}

// NewSecurityValidator creates a new SecurityValidator
func NewSecurityValidator() *SecurityValidator {
	return &SecurityValidator{
		blockedCommands: []string{
			"rm -rf /",
			"rm -rf /*",
			"mkfs",
			"dd if=/dev/zero",
			"> /dev/sda",
			"chmod -R 777 /",
			":(){ :|:& };:", // Fork bomb
		},
		blockedPatterns: []*regexp.Regexp{
			regexp.MustCompile(`rm\s+-rf\s+/[^/\s]*$`),
			regexp.MustCompile(`>\s*/etc/`),
			regexp.MustCompile(`curl.*\|\s*bash`),
			regexp.MustCompile(`wget.*\|\s*sh`),
			regexp.MustCompile(`curl.*\|\s*sh`),
			regexp.MustCompile(`wget.*\|\s*bash`),
		},
		dangerousPatterns: []string{
			"sudo",
			"rm -r",
			"rm -f",
			"systemctl",
			"service",
			"reboot",
			"shutdown",
			"kill -9",
			"pkill",
			"killall",
			"chmod",
			"chown",
			"iptables",
			"firewall",
		},
	}
}

// ValidateCommand checks if a command is safe to execute
func (s *SecurityValidator) ValidateCommand(command string) error {
	// Check blocked commands
	for _, blocked := range s.blockedCommands {
		if strings.Contains(command, blocked) {
			return fmt.Errorf("blocked command detected: %s", blocked)
		}
	}

	// Check blocked patterns
	for _, pattern := range s.blockedPatterns {
		if pattern.MatchString(command) {
			return fmt.Errorf("potentially dangerous command pattern detected")
		}
	}

	return nil
}

// RequiresConfirmation returns true if command should require user confirmation
func (s *SecurityValidator) RequiresConfirmation(command string) bool {
	cmdLower := strings.ToLower(command)

	for _, pattern := range s.dangerousPatterns {
		if strings.Contains(cmdLower, pattern) {
			return true
		}
	}

	return false
}

// GetWarnings returns warnings for potentially dangerous commands
func (s *SecurityValidator) GetWarnings(command string) []string {
	var warnings []string
	cmdLower := strings.ToLower(command)

	if strings.Contains(cmdLower, "sudo") {
		warnings = append(warnings, "Command uses sudo - will require elevated privileges")
	}

	if strings.Contains(cmdLower, "rm ") {
		warnings = append(warnings, "Command includes file removal")
	}

	if strings.Contains(cmdLower, "systemctl") || strings.Contains(cmdLower, "service") {
		warnings = append(warnings, "Command modifies system services")
	}

	if strings.Contains(cmdLower, "chmod") || strings.Contains(cmdLower, "chown") {
		warnings = append(warnings, "Command modifies file permissions")
	}

	return warnings
}
