package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/curl"
	"github.com/spf13/cobra"
)

// NewCurlCommand creates the curl command
func NewCurlCommand(usecase *curl.CurlUsecase) *cobra.Command {
	var method string
	var headerFlags []string
	var body string
	var timeout time.Duration
	var followRedirects bool
	var insecure bool
	var showBody bool

	cmd := &cobra.Command{
		Use:   "curl <url>",
		Short: "Faz requisição HTTP com análise detalhada",
		Long: `Faz uma requisição HTTP e mostra análise detalhada.

Similar ao curl, mas com:
  - Análise de timing (DNS, TCP, TLS, etc)
  - Formatação amigável
  - Análise de cookies
  - Informações de TLS`,
		Example: `  jorge curl https://api.exemplo.com
  jorge curl https://api.exemplo.com -X POST -d '{"test":true}'
  jorge curl https://api.exemplo.com -H "Authorization: Bearer token"
  jorge curl https://api.exemplo.com --show-body`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			url := args[0]

			ctx, cancel := context.WithTimeout(context.Background(), timeout+10*time.Second)
			defer cancel()

			// Parse headers
			headers := make(map[string]string)
			for _, h := range headerFlags {
				parts := strings.SplitN(h, ":", 2)
				if len(parts) == 2 {
					headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
				}
			}

			config := models.CurlRequest{
				URL:             url,
				Method:          method,
				Headers:         headers,
				Body:            body,
				Timeout:         timeout,
				FollowRedirects: followRedirects,
				Insecure:        insecure,
				ShowBody:        showBody,
			}

			fmt.Println()
			fmt.Printf("   %s %s\n", method, url)
			fmt.Println()

			result, err := usecase.Execute(ctx, config)
			if err != nil {
				return err
			}

			displayCurlResult(result)
			return nil
		},
	}

	cmd.Flags().StringVarP(&method, "method", "X", "GET", "HTTP method")
	cmd.Flags().StringArrayVarP(&headerFlags, "header", "H", nil, "Headers (pode repetir)")
	cmd.Flags().StringVarP(&body, "data", "d", "", "Request body")
	cmd.Flags().DurationVarP(&timeout, "timeout", "t", 30*time.Second, "Timeout")
	cmd.Flags().BoolVarP(&followRedirects, "follow", "L", true, "Seguir redirects")
	cmd.Flags().BoolVarP(&insecure, "insecure", "k", false, "Ignorar erros de certificado")
	cmd.Flags().BoolVar(&showBody, "show-body", false, "Mostrar corpo da resposta")

	return cmd
}

func displayCurlResult(result *models.CurlResult) {
	// Timing
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println(" TIMING")
	fmt.Println(strings.Repeat("─", 60))

	timings := []struct {
		name  string
		value time.Duration
	}{
		{"DNS Lookup", result.Timing.DNSLookup},
		{"TCP Connect", result.Timing.TCPConnect},
		{"TLS Handshake", result.Timing.TLSHandshake},
		{"Server Processing", result.Timing.ServerProcessing},
		{"Content Transfer", result.Timing.ContentTransfer},
	}

	for _, t := range timings {
		if t.value > 0 {
			fmt.Printf("   %-20s %s\n", t.name+":", t.value.Round(time.Millisecond))
		}
	}
	fmt.Printf("   %-20s %s\n",
		lipgloss.NewStyle().Bold(true).Render("Total:"),
		lipgloss.NewStyle().Bold(true).Render(result.Timing.Total.Round(time.Millisecond).String()))

	// Response
	fmt.Println()
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println(" RESPONSE")
	fmt.Println(strings.Repeat("─", 60))

	if result.Success {
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
		fmt.Printf("   Size: %d bytes\n", result.Response.BodySize)

		// Headers
		fmt.Println()
		fmt.Println("   Headers:")
		for k, v := range result.Response.Headers {
			// Skip some headers for brevity
			if k == "Set-Cookie" || k == "Date" || k == "Connection" {
				continue
			}
			value := v
			if len(value) > 50 {
				value = value[:47] + "..."
			}
			fmt.Printf("     %s: %s\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render(k),
				value)
		}

		// Cookies
		if len(result.Response.Cookies) > 0 {
			fmt.Println()
			fmt.Println("   Cookies:")
			for _, c := range result.Response.Cookies {
				flags := []string{}
				if c.Secure {
					flags = append(flags, "Secure")
				}
				if c.HttpOnly {
					flags = append(flags, "HttpOnly")
				}
				flagStr := ""
				if len(flags) > 0 {
					flagStr = " [" + strings.Join(flags, ", ") + "]"
				}
				fmt.Printf("     %s=%s%s\n", c.Name, truncateStr(c.Value, 30), flagStr)
			}
		}

		// Body
		if result.Response.Body != "" {
			fmt.Println()
			fmt.Println(strings.Repeat("─", 60))
			fmt.Println(" BODY")
			fmt.Println(strings.Repeat("─", 60))

			body := result.Response.Body
			if len(body) > 1000 {
				body = body[:1000] + "\n... (truncated)"
			}
			fmt.Println(body)
		}
	} else {
		fmt.Printf("   %s %s\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("Error:"),
			result.Error)
	}

	// TLS
	if result.TLS != nil {
		fmt.Println()
		fmt.Println(strings.Repeat("─", 60))
		fmt.Println(" TLS")
		fmt.Println(strings.Repeat("─", 60))
		fmt.Printf("   Version: %s\n", result.TLS.Version)
		fmt.Printf("   Cipher: %s\n", result.TLS.CipherSuite)
		fmt.Printf("   Certificate: %s (by %s)\n", result.TLS.Certificate, result.TLS.Issuer)
	}

	// Redirects
	if len(result.Redirects) > 0 {
		fmt.Println()
		fmt.Println(strings.Repeat("─", 60))
		fmt.Println(" REDIRECTS")
		fmt.Println(strings.Repeat("─", 60))
		for _, r := range result.Redirects {
			fmt.Printf("   → %s\n", r.To)
		}
	}

	fmt.Println()
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
