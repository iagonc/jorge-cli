package commands

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/bench"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewHTTPBenchCommand(usecase *bench.HTTPBenchmarkUsecase) *cobra.Command {
	var url string
	var method string
	var requests int
	var concurrency int
	var timeout int
	var body string
	var headers []string

	cmd := &cobra.Command{
		Use:   "http-bench",
		Short: "Simple HTTP load testing",
		Long:  "Performs HTTP load testing with configurable concurrency and provides latency percentiles (p50, p95, p99).",
		Example: `  # Basic benchmark
  jorge-cli http-bench --url https://api.example.com/health

  # Custom requests and concurrency
  jorge-cli http-bench --url https://api.example.com/health --requests 1000 --concurrency 50

  # POST request with body
  jorge-cli http-bench --url https://api.example.com/data --method POST --body '{"key":"value"}'

  # With custom headers
  jorge-cli http-bench --url https://api.example.com --header "Authorization: Bearer token"`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()

			// Parse headers
			headerMap := make(map[string]string)
			for _, h := range headers {
				var key, value string
				n, _ := fmt.Sscanf(h, "%[^:]: %s", &key, &value)
				if n >= 2 {
					headerMap[key] = value
				}
			}

			config := models.BenchmarkConfig{
				URL:         url,
				Method:      method,
				Requests:    requests,
				Concurrency: concurrency,
				Timeout:     time.Duration(timeout) * time.Second,
				Body:        body,
				Headers:     headerMap,
			}

			s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
			s.Suffix = fmt.Sprintf(" Running benchmark (%d requests, %d concurrent)...", requests, concurrency)
			s.Start()

			result, err := usecase.RunBenchmark(ctx, config)
			s.Stop()

			if err != nil {
				usecase.Logger.Error("Benchmark failed", zap.Error(err))
				errorStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF6347"))
				fmt.Println(errorStyle.Render("Error: " + err.Error()))
				return
			}

			// Display results
			displayBenchmarkResult(result)
		},
	}

	cmd.Flags().StringVarP(&url, "url", "u", "", "URL to benchmark (required)")
	cmd.Flags().StringVarP(&method, "method", "m", "GET", "HTTP method")
	cmd.Flags().IntVarP(&requests, "requests", "n", 100, "Number of requests")
	cmd.Flags().IntVarP(&concurrency, "concurrency", "c", 10, "Number of concurrent workers")
	cmd.Flags().IntVarP(&timeout, "timeout", "t", 30, "Request timeout in seconds")
	cmd.Flags().StringVarP(&body, "body", "b", "", "Request body")
	cmd.Flags().StringArrayVarP(&headers, "header", "H", nil, "Headers (format: Key: Value)")
	cmd.MarkFlagRequired("url")

	return cmd
}

func displayBenchmarkResult(result *models.BenchmarkResult) {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	successStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#10B981"))
	errorStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#EF4444"))
	highlightStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F59E0B"))

	fmt.Println()
	fmt.Println(titleStyle.Render("HTTP Benchmark Results"))
	fmt.Println(titleStyle.Render("======================"))
	fmt.Println()

	fmt.Printf("  %s %s\n", labelStyle.Render("URL:"), result.URL)
	fmt.Println()

	// Request summary
	fmt.Println(titleStyle.Render("Requests:"))
	fmt.Printf("  %s %d\n", labelStyle.Render("Total:"), result.TotalRequests)
	fmt.Printf("  %s %s\n", labelStyle.Render("Successful:"), successStyle.Render(fmt.Sprintf("%d", result.SuccessfulRequests)))
	fmt.Printf("  %s %s\n", labelStyle.Render("Failed:"), errorStyle.Render(fmt.Sprintf("%d", result.FailedRequests)))
	fmt.Println()

	// Performance
	fmt.Println(titleStyle.Render("Performance:"))
	fmt.Printf("  %s %s\n", labelStyle.Render("Duration:"), result.TotalDuration.Round(time.Millisecond).String())
	fmt.Printf("  %s %s\n", labelStyle.Render("Requests/sec:"), highlightStyle.Render(fmt.Sprintf("%.2f", result.RequestsPerSecond)))
	fmt.Printf("  %s %s\n", labelStyle.Render("Data received:"), formatBytes(result.BytesReceived))
	fmt.Println()

	// Latency
	fmt.Println(titleStyle.Render("Latency:"))
	fmt.Printf("  %s %s\n", labelStyle.Render("Min:"), result.Latency.Min.Round(time.Microsecond).String())
	fmt.Printf("  %s %s\n", labelStyle.Render("Max:"), result.Latency.Max.Round(time.Microsecond).String())
	fmt.Printf("  %s %s\n", labelStyle.Render("Mean:"), result.Latency.Mean.Round(time.Microsecond).String())
	fmt.Printf("  %s %s\n", labelStyle.Render("P50:"), highlightStyle.Render(result.Latency.P50.Round(time.Microsecond).String()))
	fmt.Printf("  %s %s\n", labelStyle.Render("P95:"), highlightStyle.Render(result.Latency.P95.Round(time.Microsecond).String()))
	fmt.Printf("  %s %s\n", labelStyle.Render("P99:"), highlightStyle.Render(result.Latency.P99.Round(time.Microsecond).String()))
	fmt.Println()

	// Status codes
	if len(result.StatusCodes) > 0 {
		fmt.Println(titleStyle.Render("Status Codes:"))
		for code, count := range result.StatusCodes {
			var codeStyle lipgloss.Style
			if code >= 200 && code < 300 {
				codeStyle = successStyle
			} else if code >= 400 {
				codeStyle = errorStyle
			} else {
				codeStyle = highlightStyle
			}
			fmt.Printf("  %s: %d\n", codeStyle.Render(fmt.Sprintf("%d", code)), count)
		}
		fmt.Println()
	}

	// Errors
	if len(result.Errors) > 0 {
		fmt.Println(errorStyle.Render("Errors:"))
		for errMsg, count := range result.Errors {
			fmt.Printf("  %s: %d\n", errMsg, count)
		}
		fmt.Println()
	}
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
