package commands

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/health"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewHealthCheckCommand(usecase *health.HealthCheckUsecase) *cobra.Command {
	var configPath string
	var url string
	var parallel bool
	var expectedStatus int

	cmd := &cobra.Command{
		Use:   "health-check",
		Short: "Perform health checks on endpoints",
		Long:  "Performs health checks on multiple endpoints defined in a YAML configuration file or a single URL.",
		Example: `  # Check from config file
  jorge-cli health-check --config endpoints.yaml

  # Check single URL
  jorge-cli health-check --url https://api.example.com/health

  # Check with expected status
  jorge-cli health-check --url https://api.example.com/health --expected-status 204`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()

			s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
			s.Suffix = " Running health checks..."
			s.Start()

			var summary *models.HealthCheckSummary
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

				summary, errorsList = usecase.RunHealthChecks(ctx, config, parallel)
			} else if url != "" {
				result, err := usecase.CheckSingle(ctx, url, expectedStatus)
				if err != nil {
					errorsList = append(errorsList, err)
				}

				summary = &models.HealthCheckSummary{
					TotalChecks: 1,
					Results:     []models.HealthCheckResult{*result},
				}
				if result.Status == models.StatusHealthy {
					summary.Healthy = 1
				} else if result.Status == models.StatusUnhealthy {
					summary.Unhealthy = 1
				} else {
					summary.Errors = 1
				}
			} else {
				s.Stop()
				errorStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF6347"))
				fmt.Println(errorStyle.Render("Error: Either --config or --url must be provided"))
				return
			}

			s.Stop()

			// Display results
			displayHealthCheckResult(summary)

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
	cmd.Flags().StringVarP(&url, "url", "u", "", "Single URL to check")
	cmd.Flags().BoolVarP(&parallel, "parallel", "P", true, "Run checks in parallel")
	cmd.Flags().IntVar(&expectedStatus, "expected-status", 200, "Expected HTTP status code")

	return cmd
}

func displayHealthCheckResult(summary *models.HealthCheckSummary) {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	successStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#10B981"))
	errorStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#EF4444"))
	warningStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F59E0B"))

	fmt.Println()
	fmt.Println(titleStyle.Render("Health Check Results"))
	fmt.Println(titleStyle.Render("===================="))
	fmt.Println()

	// Summary
	fmt.Println(titleStyle.Render("Summary:"))
	fmt.Printf("  %s %d\n", labelStyle.Render("Total Checks:"), summary.TotalChecks)
	fmt.Printf("  %s %s\n", labelStyle.Render("Healthy:"), successStyle.Render(fmt.Sprintf("%d", summary.Healthy)))
	fmt.Printf("  %s %s\n", labelStyle.Render("Unhealthy:"), errorStyle.Render(fmt.Sprintf("%d", summary.Unhealthy)))
	fmt.Printf("  %s %s\n", labelStyle.Render("Errors:"), warningStyle.Render(fmt.Sprintf("%d", summary.Errors)))
	fmt.Printf("  %s %s\n", labelStyle.Render("Duration:"), summary.Duration.Round(time.Millisecond).String())
	fmt.Println()

	// Detailed results
	fmt.Println(titleStyle.Render("Details:"))

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	fmt.Println(headerStyle.Render(fmt.Sprintf("%-25s %-10s %-8s %-12s", "NAME", "STATUS", "CODE", "RESPONSE")))

	for _, result := range summary.Results {
		var statusStr string
		switch result.Status {
		case models.StatusHealthy:
			statusStr = successStyle.Render("healthy")
		case models.StatusUnhealthy:
			statusStr = errorStyle.Render("unhealthy")
		case models.StatusError:
			statusStr = warningStyle.Render("error")
		}

		name := result.Name
		if len(name) > 24 {
			name = name[:21] + "..."
		}

		codeStr := fmt.Sprintf("%d", result.StatusCode)
		if result.StatusCode == 0 {
			codeStr = "-"
		}

		responseTime := result.ResponseTime.Round(time.Millisecond).String()

		fmt.Printf("  %-25s %-10s %-8s %-12s\n", name, statusStr, codeStr, responseTime)

		if result.Error != "" {
			fmt.Printf("    %s\n", warningStyle.Render("Error: "+result.Error))
		}
	}
	fmt.Println()
}
