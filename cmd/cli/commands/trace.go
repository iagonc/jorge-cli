package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/trace"
	"github.com/spf13/cobra"
)

// NewTraceCommand creates the trace command
func NewTraceCommand(usecase *trace.TraceUsecase) *cobra.Command {
	var method string
	var headers []string
	var body string
	var showBody bool

	cmd := &cobra.Command{
		Use:   "trace <url>",
		Short: "Rastreia cada fase de uma requisição HTTP",
		Long: `Mostra o tempo de cada fase de uma requisição HTTP:
  - DNS Lookup
  - TCP Connect
  - TLS Handshake
  - Server Processing (Time to First Byte)
  - Content Transfer

Útil para identificar onde está a lentidão.`,
		Example: `  jorge trace https://api.exemplo.com
  jorge trace https://api.exemplo.com -X POST -d '{"test":true}'
  jorge trace https://api.exemplo.com -H "Authorization: Bearer token"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			url := args[0]

			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			headerMap := make(map[string]string)
			for _, h := range headers {
				parts := strings.SplitN(h, ":", 2)
				if len(parts) == 2 {
					headerMap[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
				}
			}

			fmt.Println()
			fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("⚡ HTTP Trace"))
			fmt.Printf("   %s %s\n\n", method, url)

			result, err := usecase.Trace(ctx, url, method, headerMap, body, showBody)
			if err != nil {
				return err
			}

			displayTraceResult(result)
			return nil
		},
	}

	cmd.Flags().StringVarP(&method, "method", "X", "GET", "HTTP method")
	cmd.Flags().StringArrayVarP(&headers, "header", "H", nil, "Headers (pode repetir)")
	cmd.Flags().StringVarP(&body, "data", "d", "", "Request body")
	cmd.Flags().BoolVar(&showBody, "show-body", false, "Mostrar corpo da resposta")

	return cmd
}

func displayTraceResult(result *models.TraceResult) {
	// Timing diagram
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println(" FASES DA REQUISIÇÃO")
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println()

	maxWidth := 40
	totalMs := float64(result.TotalDuration.Milliseconds())
	if totalMs == 0 {
		totalMs = 1
	}

	for _, phase := range result.Phases {
		phaseMs := float64(phase.Duration.Milliseconds())
		barWidth := int((phaseMs / totalMs) * float64(maxWidth))
		if barWidth < 1 && phase.Duration > 0 {
			barWidth = 1
		}

		var statusIcon string
		var barColor lipgloss.Color
		if phase.Success {
			statusIcon = "✓"
			barColor = lipgloss.Color("42")
		} else {
			statusIcon = "✗"
			barColor = lipgloss.Color("196")
		}

		bar := lipgloss.NewStyle().Foreground(barColor).Render(strings.Repeat("█", barWidth))
		timeStr := fmt.Sprintf("%6s", phase.Duration.Round(time.Millisecond))

		fmt.Printf("   %s %-20s %s %s\n", statusIcon, phase.Name, bar, timeStr)
		if phase.Details != "" {
			fmt.Printf("     %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(phase.Details))
		}
	}

	// Total
	fmt.Println()
	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("   TOTAL: %s\n", lipgloss.NewStyle().Bold(true).Render(result.TotalDuration.Round(time.Millisecond).String()))
	fmt.Println(strings.Repeat("─", 60))

	// TLS info
	if result.TLS != nil {
		fmt.Println()
		fmt.Println(" TLS INFO")
		fmt.Printf("   Version: %s | Cipher: %s\n", result.TLS.Version, result.TLS.CipherSuite)
		fmt.Printf("   Certificate: %s (by %s)\n", result.TLS.CertificateSubject, result.TLS.CertificateIssuer)
		fmt.Printf("   Expires: %s\n", result.TLS.CertificateExpiry.Format("02/01/2006"))
	}

	// Response
	if result.Response != nil {
		fmt.Println()
		fmt.Println(" RESPONSE")
		statusStyle := lipgloss.NewStyle().Bold(true)
		if result.Response.StatusCode >= 400 {
			statusStyle = statusStyle.Foreground(lipgloss.Color("196"))
		} else if result.Response.StatusCode >= 300 {
			statusStyle = statusStyle.Foreground(lipgloss.Color("214"))
		} else {
			statusStyle = statusStyle.Foreground(lipgloss.Color("42"))
		}
		fmt.Printf("   Status: %s\n", statusStyle.Render(result.Response.Status))
		fmt.Printf("   Content-Type: %s\n", result.Response.ContentType)
		fmt.Printf("   Size: %d bytes\n", result.Response.ContentLength)

		if result.Response.Body != "" {
			fmt.Println()
			fmt.Println(" BODY")
			fmt.Println(strings.Repeat("─", 60))
			// Limit body display
			body := result.Response.Body
			if len(body) > 500 {
				body = body[:500] + "..."
			}
			fmt.Println(body)
		}
	}

	// Redirects
	if len(result.Redirects) > 0 {
		fmt.Println()
		fmt.Println(" REDIRECTS")
		for _, r := range result.Redirects {
			fmt.Printf("   → %s\n", r.To)
		}
	}

	// Error
	if result.Error != "" {
		fmt.Println()
		fmt.Printf("   %s %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("ERROR:"), result.Error)
	}

	fmt.Println()
}
