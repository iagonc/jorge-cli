package commands

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/mock"
	"github.com/spf13/cobra"
)

// NewMockCommand creates the mock command
func NewMockCommand(usecase *mock.MockUsecase) *cobra.Command {
	var port int
	var host string
	var cors bool
	var defaultStatus int

	cmd := &cobra.Command{
		Use:   "mock",
		Short: "Inicia um servidor mock para testes",
		Long: `Inicia um servidor HTTP mock para testar webhooks e integrações.

Todas as requisições são logadas e respondidas com JSON.
Útil para testar webhooks, callbacks e integrações.`,
		Example: `  jorge mock
  jorge mock --port 3000
  jorge mock --port 8080 --cors`,
		RunE: func(cmd *cobra.Command, args []string) error {
			config := models.MockConfig{
				Port:          port,
				Host:          host,
				DefaultStatus: defaultStatus,
				CORS:          cors,
				LogRequests:   true,
			}

			ctx, cancel := context.WithCancel(context.Background())

			// Handle Ctrl+C
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				<-sigChan
				fmt.Println("\n\nParando servidor...")
				cancel()
			}()

			fmt.Println()
			fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("🎭 Mock Server"))
			fmt.Printf("   Listening on http://%s:%d\n", host, port)
			if cors {
				fmt.Println("   CORS: habilitado")
			}
			fmt.Println("   Pressione Ctrl+C para parar")
			fmt.Println()
			fmt.Println(strings.Repeat("─", 80))
			fmt.Printf(" %-8s %-6s %-30s %-10s %s\n", "TIME", "METHOD", "PATH", "STATUS", "SIZE")
			fmt.Println(strings.Repeat("─", 80))

			err := usecase.Start(ctx, config, func(req models.MockRequest, resp models.MockResponse) {
				displayMockRequest(req, resp)
			})

			if err != nil && err != context.Canceled {
				return err
			}

			// Print stats
			stats := usecase.GetStats()
			fmt.Println()
			fmt.Println(strings.Repeat("─", 80))
			fmt.Printf(" Total requests: %d | Uptime: %s\n",
				stats.TotalRequests,
				time.Since(stats.StartTime).Round(time.Second))
			fmt.Println()

			return nil
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 8080, "Porta do servidor")
	cmd.Flags().StringVar(&host, "host", "0.0.0.0", "Host para bind")
	cmd.Flags().BoolVar(&cors, "cors", false, "Habilitar CORS")
	cmd.Flags().IntVar(&defaultStatus, "status", 200, "Status HTTP padrão")

	return cmd
}

func displayMockRequest(req models.MockRequest, resp models.MockResponse) {
	timeStr := req.Timestamp.Format("15:04:05")

	methodStyle := lipgloss.NewStyle().Width(6)
	switch req.Method {
	case "GET":
		methodStyle = methodStyle.Foreground(lipgloss.Color("42"))
	case "POST":
		methodStyle = methodStyle.Foreground(lipgloss.Color("39"))
	case "PUT":
		methodStyle = methodStyle.Foreground(lipgloss.Color("214"))
	case "DELETE":
		methodStyle = methodStyle.Foreground(lipgloss.Color("196"))
	default:
		methodStyle = methodStyle.Foreground(lipgloss.Color("245"))
	}

	statusStyle := lipgloss.NewStyle()
	if resp.StatusCode >= 400 {
		statusStyle = statusStyle.Foreground(lipgloss.Color("196"))
	} else if resp.StatusCode >= 300 {
		statusStyle = statusStyle.Foreground(lipgloss.Color("214"))
	} else {
		statusStyle = statusStyle.Foreground(lipgloss.Color("42"))
	}

	path := req.Path
	if req.Query != "" {
		path += "?" + req.Query
	}
	if len(path) > 30 {
		path = path[:27] + "..."
	}

	sizeStr := fmt.Sprintf("%d B", req.BodySize)

	fmt.Printf(" %-8s %s %-30s %s %s\n",
		timeStr,
		methodStyle.Render(req.Method),
		path,
		statusStyle.Render(fmt.Sprintf("%-10d", resp.StatusCode)),
		sizeStr)

	// Show body preview for POST/PUT
	if req.Body != "" && (req.Method == "POST" || req.Method == "PUT" || req.Method == "PATCH") {
		body := req.Body
		if len(body) > 60 {
			body = body[:57] + "..."
		}
		fmt.Printf("          %s\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("└─ "+body))
	}
}
