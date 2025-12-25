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
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/watch"
	"github.com/spf13/cobra"
)

// NewWatchCommand creates the watch command
func NewWatchCommand(usecase *watch.WatchUsecase) *cobra.Command {
	var interval time.Duration
	var timeout time.Duration
	var expectedStatus int
	var maxFailures int

	cmd := &cobra.Command{
		Use:   "watch <url>",
		Short: "Monitora um endpoint continuamente",
		Long: `Monitora um endpoint e mostra status em tempo real.

Mostra:
  - Status atual (UP/DOWN/DEGRADED)
  - Tempo de resposta
  - Estatísticas (uptime, avg latency)
  - Alertas quando o status muda

Use Ctrl+C para parar o monitoramento.`,
		Example: `  jorge watch https://api.exemplo.com
  jorge watch https://api.exemplo.com --interval 10s
  jorge watch https://api.exemplo.com --expected-status 204
  jorge watch https://api.exemplo.com --max-failures 3`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			url := args[0]

			config := models.WatchConfig{
				URL:            url,
				Interval:       interval,
				Timeout:        timeout,
				ExpectedStatus: expectedStatus,
				MaxFailures:    maxFailures,
			}

			ctx, cancel := context.WithCancel(context.Background())

			// Handle Ctrl+C
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				<-sigChan
				fmt.Println("\n\nParando monitoramento...")
				cancel()
			}()

			fmt.Println()
			fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("👁  Watch Mode"))
			fmt.Printf("   Monitorando: %s\n", url)
			fmt.Printf("   Intervalo: %s | Timeout: %s\n", interval, timeout)
			fmt.Println("   Pressione Ctrl+C para parar")
			fmt.Println()
			fmt.Println(strings.Repeat("─", 70))

			var lastStatus models.WatchStatus

			err := usecase.Watch(ctx, config, func(event models.WatchEvent, stats models.WatchStats) {
				displayWatchEvent(event, stats, &lastStatus)
			})

			if err != nil && err != context.Canceled {
				return err
			}

			return nil
		},
	}

	cmd.Flags().DurationVarP(&interval, "interval", "i", 5*time.Second, "Intervalo entre checks")
	cmd.Flags().DurationVarP(&timeout, "timeout", "t", 10*time.Second, "Timeout por requisição")
	cmd.Flags().IntVar(&expectedStatus, "expected-status", 200, "Status HTTP esperado")
	cmd.Flags().IntVar(&maxFailures, "max-failures", 0, "Parar após N falhas consecutivas (0=infinito)")

	return cmd
}

func displayWatchEvent(event models.WatchEvent, stats models.WatchStats, lastStatus *models.WatchStatus) {
	timestamp := event.Timestamp.Format("15:04:05")

	var statusIcon, statusText string
	var statusColor lipgloss.Color

	switch event.Status {
	case models.WatchStatusUp:
		statusIcon = "●"
		statusText = "UP"
		statusColor = lipgloss.Color("42")
	case models.WatchStatusDown:
		statusIcon = "●"
		statusText = "DOWN"
		statusColor = lipgloss.Color("196")
	case models.WatchStatusDegraded:
		statusIcon = "●"
		statusText = "SLOW"
		statusColor = lipgloss.Color("214")
	}

	statusStyle := lipgloss.NewStyle().Foreground(statusColor).Bold(true)

	// Show status change alert
	if *lastStatus != "" && *lastStatus != event.Status {
		if event.Status == models.WatchStatusDown {
			fmt.Printf("\n   %s ALERTA: Endpoint caiu! %s\n\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("⚠"),
				event.Error)
		} else if *lastStatus == models.WatchStatusDown && event.Status == models.WatchStatusUp {
			fmt.Printf("\n   %s Endpoint recuperou!\n\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓"))
		}
	}
	*lastStatus = event.Status

	// Main status line
	latencyStr := event.ResponseTime.Round(time.Millisecond).String()
	fmt.Printf("   %s %s %s  %s  ",
		timestamp,
		statusStyle.Render(statusIcon),
		statusStyle.Render(fmt.Sprintf("%-4s", statusText)),
		fmt.Sprintf("%8s", latencyStr))

	// Stats
	uptimeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	if stats.Uptime >= 99 {
		uptimeStyle = uptimeStyle.Foreground(lipgloss.Color("42"))
	} else if stats.Uptime < 90 {
		uptimeStyle = uptimeStyle.Foreground(lipgloss.Color("196"))
	}

	fmt.Printf("│ Uptime: %s  Avg: %s  Checks: %d/%d",
		uptimeStyle.Render(fmt.Sprintf("%.1f%%", stats.Uptime)),
		stats.AvgResponseTime.Round(time.Millisecond),
		stats.SuccessfulChecks,
		stats.TotalChecks)

	if event.Error != "" && event.Status == models.WatchStatusDown {
		fmt.Printf("  │ %s", lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(truncateWatch(event.Error, 30)))
	}

	fmt.Println()
}

func truncateWatch(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
