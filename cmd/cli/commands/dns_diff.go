package commands

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/dns"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewDNSDiffCommand(usecase *dns.DNSDiffUsecase) *cobra.Command {
	var configPath string
	var domain string
	var recordType string
	var expected string
	var nameserver string
	var showAll bool

	cmd := &cobra.Command{
		Use:   "dns-diff",
		Short: "Compare DNS records against expected values",
		Long:  "Compares actual DNS records with expected values to detect configuration drift or propagation issues.",
		Example: `  # Check from config file
  jorge-cli dns-diff --config dns.yaml

  # Check single record
  jorge-cli dns-diff --domain example.com --type A --expected 93.184.216.34

  # Use custom nameserver
  jorge-cli dns-diff --config dns.yaml --nameserver 8.8.8.8`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()

			// Use custom nameserver if provided
			if nameserver != "" {
				usecase = dns.NewDNSDiffUsecaseWithNameserver(usecase.Logger, nameserver)
			}

			s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
			s.Suffix = " Comparing DNS records..."
			s.Start()

			var result *models.DNSDiffResult
			var errorsList []error

			if configPath != "" {
				config, err := usecase.LoadConfig(configPath)
				if err != nil {
					s.Stop()
					usecase.Logger.Error("Failed to load config", zap.Error(err))
					errorStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF6347"))
					fmt.Println(errorStyle.Render("Error: " + err.Error()))
					return
				}

				if nameserver != "" {
					config.Nameserver = nameserver
				}

				result, errorsList = usecase.CompareDNS(ctx, config)
			} else if domain != "" && expected != "" {
				diff, err := usecase.CheckSingleRecord(ctx, domain, models.DNSRecordType(recordType), expected)
				if err != nil {
					errorsList = append(errorsList, err)
				}

				result = &models.DNSDiffResult{
					Nameserver:   nameserver,
					TotalRecords: 1,
					Diffs:        []models.RecordDiff{*diff},
				}
				if diff.Match {
					result.Matching = 1
				} else if diff.Error != "" {
					result.Errors = 1
				} else {
					result.Mismatched = 1
				}
			} else {
				s.Stop()
				errorStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF6347"))
				fmt.Println(errorStyle.Render("Error: Either --config or (--domain and --expected) must be provided"))
				return
			}

			s.Stop()

			// Display results
			displayDNSDiffResult(result, showAll)

			// Display errors if any
			if len(errorsList) > 0 {
				fmt.Println()
				errorStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF6347"))
				fmt.Println(errorStyle.Render("Errors encountered:"))
				for _, err := range errorsList {
					fmt.Printf("  - %v\n", err)
				}
			}
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to YAML config file")
	cmd.Flags().StringVarP(&domain, "domain", "d", "", "Domain to check")
	cmd.Flags().StringVarP(&recordType, "type", "T", "A", "Record type (A, AAAA, CNAME, MX, TXT, NS)")
	cmd.Flags().StringVarP(&expected, "expected", "e", "", "Expected value")
	cmd.Flags().StringVarP(&nameserver, "nameserver", "n", "", "Custom DNS server (e.g., 8.8.8.8)")
	cmd.Flags().BoolVarP(&showAll, "all", "a", false, "Show all records, not just differences")

	return cmd
}

func displayDNSDiffResult(result *models.DNSDiffResult, showAll bool) {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	successStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#10B981"))
	errorStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#EF4444"))
	warningStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F59E0B"))

	fmt.Println()
	fmt.Println(titleStyle.Render("DNS Diff Results"))
	fmt.Println(titleStyle.Render("================"))
	fmt.Println()

	if result.Nameserver != "" {
		fmt.Printf("  %s %s\n", labelStyle.Render("Nameserver:"), result.Nameserver)
	}
	fmt.Printf("  %s %s\n", labelStyle.Render("Duration:"), result.Duration.Round(time.Millisecond).String())
	fmt.Println()

	// Summary
	fmt.Println(titleStyle.Render("Summary:"))
	fmt.Printf("  %s %d\n", labelStyle.Render("Total Records:"), result.TotalRecords)
	fmt.Printf("  %s %s\n", labelStyle.Render("Matching:"), successStyle.Render(fmt.Sprintf("%d", result.Matching)))
	fmt.Printf("  %s %s\n", labelStyle.Render("Mismatched:"), errorStyle.Render(fmt.Sprintf("%d", result.Mismatched)))
	fmt.Printf("  %s %s\n", labelStyle.Render("Errors:"), warningStyle.Render(fmt.Sprintf("%d", result.Errors)))
	fmt.Println()

	// Detailed results
	hasDifferences := false
	for _, diff := range result.Diffs {
		if !diff.Match || showAll {
			if !hasDifferences {
				fmt.Println(titleStyle.Render("Details:"))
				hasDifferences = true
			}

			var statusStr string
			if diff.Error != "" {
				statusStr = warningStyle.Render("[ERROR]")
			} else if diff.Match {
				statusStr = successStyle.Render("[MATCH]")
			} else {
				statusStr = errorStyle.Render("[MISMATCH]")
			}

			fmt.Printf("\n  %s %s %s\n", statusStr, diff.Domain, labelStyle.Render(string(diff.RecordType)))

			if diff.Error != "" {
				fmt.Printf("    %s %s\n", labelStyle.Render("Error:"), diff.Error)
				continue
			}

			fmt.Printf("    %s %v\n", labelStyle.Render("Expected:"), diff.Expected)
			fmt.Printf("    %s %v\n", labelStyle.Render("Actual:"), diff.Actual)

			if len(diff.Missing) > 0 {
				fmt.Printf("    %s %v\n", errorStyle.Render("Missing:"), diff.Missing)
			}
			if len(diff.Extra) > 0 {
				fmt.Printf("    %s %v\n", warningStyle.Render("Extra:"), diff.Extra)
			}
		}
	}

	if !hasDifferences && !showAll {
		fmt.Println(successStyle.Render("All DNS records match expected values!"))
	}
	fmt.Println()
}
