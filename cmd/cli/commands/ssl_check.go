package commands

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/ssl"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewSSLCheckCommand(usecase *ssl.SSLCheckUsecase) *cobra.Command {
	var host string
	var port int

	cmd := &cobra.Command{
		Use:   "ssl-check",
		Short: "Check SSL/TLS certificate for a host",
		Long:  "Performs a comprehensive SSL/TLS certificate check including validity, expiration, issuer, and certificate chain.",
		Example: `  # Check SSL certificate on default port 443
  jorge-cli ssl-check --host example.com

  # Check SSL certificate on custom port
  jorge-cli ssl-check --host example.com --port 8443`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()

			s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
			s.Suffix = " Checking SSL certificate..."
			s.Start()

			result, err := usecase.CheckSSL(ctx, host, port)
			s.Stop()

			if err != nil {
				usecase.Logger.Error("SSL check failed", zap.Error(err))
				errorStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF6347"))
				fmt.Println(errorStyle.Render("Error: " + err.Error()))
				return
			}

			// Display results
			displaySSLResult(result)
		},
	}

	cmd.Flags().StringVarP(&host, "host", "H", "", "Host to check SSL certificate (required)")
	cmd.Flags().IntVarP(&port, "port", "p", 443, "Port to connect to")
	cmd.MarkFlagRequired("host")

	return cmd
}

func displaySSLResult(result *models.SSLCheckResult) {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	successStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#10B981"))
	warningStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F59E0B"))
	errorStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#EF4444"))

	fmt.Println()
	fmt.Println(titleStyle.Render("SSL Certificate Check Results"))
	fmt.Println(titleStyle.Render("============================="))
	fmt.Println()

	fmt.Println(titleStyle.Render("Connection Info:"))
	fmt.Printf("  %s %s:%d\n", labelStyle.Render("Host:"), result.Host, result.Port)
	fmt.Printf("  %s %s\n", labelStyle.Render("TLS Version:"), result.TLSVersion)
	fmt.Printf("  %s %s\n", labelStyle.Render("Cipher Suite:"), result.CipherSuite)
	fmt.Println()

	fmt.Println(titleStyle.Render("Certificate Info:"))
	fmt.Printf("  %s %s\n", labelStyle.Render("Subject:"), result.Certificate.Subject)
	fmt.Printf("  %s %s\n", labelStyle.Render("Issuer:"), result.Certificate.Issuer)
	fmt.Printf("  %s %s\n", labelStyle.Render("Valid From:"), result.Certificate.ValidFrom.Format(time.RFC3339))
	fmt.Printf("  %s %s\n", labelStyle.Render("Valid Until:"), result.Certificate.ValidUntil.Format(time.RFC3339))
	fmt.Printf("  %s %s\n", labelStyle.Render("Serial Number:"), result.Certificate.SerialNumber)
	fmt.Printf("  %s %s\n", labelStyle.Render("Signature Algorithm:"), result.Certificate.SignatureAlgo)

	if len(result.Certificate.DNSNames) > 0 {
		fmt.Printf("  %s\n", labelStyle.Render("DNS Names:"))
		for _, name := range result.Certificate.DNSNames {
			fmt.Printf("    - %s\n", name)
		}
	}
	fmt.Println()

	// Expiry status
	fmt.Println(titleStyle.Render("Status:"))
	if result.Certificate.IsExpired {
		fmt.Printf("  %s %s\n", labelStyle.Render("Validity:"), errorStyle.Render("EXPIRED"))
	} else if result.Certificate.DaysUntilExpiry <= 7 {
		fmt.Printf("  %s %s\n", labelStyle.Render("Validity:"), errorStyle.Render(fmt.Sprintf("EXPIRES IN %d DAYS", result.Certificate.DaysUntilExpiry)))
	} else if result.Certificate.DaysUntilExpiry <= 30 {
		fmt.Printf("  %s %s\n", labelStyle.Render("Validity:"), warningStyle.Render(fmt.Sprintf("Expires in %d days", result.Certificate.DaysUntilExpiry)))
	} else {
		fmt.Printf("  %s %s\n", labelStyle.Render("Validity:"), successStyle.Render(fmt.Sprintf("Valid (%d days remaining)", result.Certificate.DaysUntilExpiry)))
	}

	if result.IsChainValid {
		fmt.Printf("  %s %s\n", labelStyle.Render("Chain:"), successStyle.Render("Valid"))
	} else {
		fmt.Printf("  %s %s\n", labelStyle.Render("Chain:"), warningStyle.Render("Issues detected"))
	}
	fmt.Println()

	// Certificate chain
	if len(result.Chain) > 0 {
		fmt.Println(titleStyle.Render("Certificate Chain:"))
		for _, cert := range result.Chain {
			level := ""
			if cert.Level == 0 {
				level = "Leaf"
			} else if cert.IsCA {
				level = fmt.Sprintf("CA (Level %d)", cert.Level)
			} else {
				level = fmt.Sprintf("Intermediate (Level %d)", cert.Level)
			}
			fmt.Printf("  [%s] %s\n", level, cert.Subject)
		}
		fmt.Println()
	}

	// Warnings
	if len(result.Warnings) > 0 {
		fmt.Println(warningStyle.Render("Warnings:"))
		for _, w := range result.Warnings {
			fmt.Printf("  - %s\n", w)
		}
		fmt.Println()
	}

	// Errors
	if len(result.Errors) > 0 {
		fmt.Println(errorStyle.Render("Errors:"))
		for _, e := range result.Errors {
			fmt.Printf("  - %s\n", e)
		}
	}
}
