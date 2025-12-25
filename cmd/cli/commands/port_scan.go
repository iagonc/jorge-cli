package commands

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/portscan"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewPortScanCommand(usecase *portscan.PortScanUsecase) *cobra.Command {
	var host string
	var portRange string
	var timeout int
	var concurrency int

	cmd := &cobra.Command{
		Use:   "port-scan",
		Short: "Scan ports on a host",
		Long:  "Scans specified ports on a host to identify open services.",
		Example: `  # Scan common ports
  jorge-cli port-scan --host 10.0.0.1 --ports 22,80,443,8080

  # Scan port range
  jorge-cli port-scan --host example.com --ports 1-1000

  # Scan with custom timeout and concurrency
  jorge-cli port-scan --host 10.0.0.1 --ports 1-65535 --timeout 1 --concurrency 200`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()

			// Parse ports
			ports, err := portscan.ParsePortRange(portRange)
			if err != nil {
				usecase.Logger.Error("Invalid port range", zap.Error(err))
				errorStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF6347"))
				fmt.Println(errorStyle.Render("Error: " + err.Error()))
				return
			}

			// Set options
			usecase.Timeout = time.Duration(timeout) * time.Second
			usecase.Concurrency = concurrency

			s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
			s.Suffix = fmt.Sprintf(" Scanning %d ports on %s...", len(ports), host)
			s.Start()

			result, errorsList := usecase.ScanPorts(ctx, host, ports)
			s.Stop()

			// Display results
			displayPortScanResult(result)

			// Display errors if any
			if len(errorsList) > 0 {
				errorStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF6347"))
				fmt.Println(errorStyle.Render("Errors:"))
				for _, err := range errorsList {
					fmt.Printf("  - %v\n", err)
				}
			}
		},
	}

	cmd.Flags().StringVarP(&host, "host", "H", "", "Host to scan (required)")
	cmd.Flags().StringVarP(&portRange, "ports", "p", "1-1000", "Port range (e.g., 1-1000, 80,443,8080)")
	cmd.Flags().IntVarP(&timeout, "timeout", "t", 2, "Timeout per port in seconds")
	cmd.Flags().IntVarP(&concurrency, "concurrency", "c", 100, "Number of concurrent scans")
	cmd.MarkFlagRequired("host")

	return cmd
}

func displayPortScanResult(result *models.PortScanResult) {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	successStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#10B981"))
	closedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))

	fmt.Println()
	fmt.Println(titleStyle.Render("Port Scan Results"))
	fmt.Println(titleStyle.Render("================="))
	fmt.Println()

	fmt.Printf("  %s %s\n", labelStyle.Render("Host:"), result.Host)
	fmt.Printf("  %s %d\n", labelStyle.Render("Ports Scanned:"), result.TotalPorts)
	fmt.Printf("  %s %s\n", labelStyle.Render("Duration:"), result.Duration.Round(time.Millisecond))
	fmt.Println()

	// Summary
	fmt.Println(titleStyle.Render("Summary:"))
	fmt.Printf("  %s %s\n", labelStyle.Render("Open:"), successStyle.Render(fmt.Sprintf("%d", len(result.OpenPorts))))
	fmt.Printf("  %s %s\n", labelStyle.Render("Closed:"), closedStyle.Render(fmt.Sprintf("%d", result.ClosedPorts)))
	fmt.Printf("  %s %s\n", labelStyle.Render("Filtered:"), closedStyle.Render(fmt.Sprintf("%d", result.FilteredPorts)))
	fmt.Println()

	// Open ports
	if len(result.OpenPorts) > 0 {
		fmt.Println(titleStyle.Render("Open Ports:"))

		headerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

		fmt.Println(headerStyle.Render(fmt.Sprintf("%-8s %-15s %-10s", "PORT", "SERVICE", "STATUS")))

		for _, port := range result.OpenPorts {
			service := port.Service
			if service == "" {
				service = "unknown"
			}
			fmt.Printf("  %-8d %-15s %s\n", port.Port, service, successStyle.Render("open"))
		}
	} else {
		fmt.Println(closedStyle.Render("No open ports found"))
	}
	fmt.Println()
}
