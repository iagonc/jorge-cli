package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/mtr"
	"github.com/spf13/cobra"
)

// NewMTRCommand creates the mtr command
func NewMTRCommand(usecase *mtr.MTRUsecase) *cobra.Command {
	var count int

	cmd := &cobra.Command{
		Use:   "mtr <host>",
		Short: "Analisa o caminho de rede até um host",
		Long: `Combina traceroute com ping para análise detalhada da rota de rede.

Mostra cada hop (roteador) no caminho até o destino, com:
  - Latência média, mínima e máxima
  - Perda de pacotes
  - Identificação de gargalos

Útil para identificar onde está o problema de conectividade.`,
		Example: `  jorge mtr google.com
  jorge mtr 8.8.8.8 --count 20`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			host := args[0]

			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			defer cancel()

			fmt.Println()
			fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("🛤  Network Path Analysis"))
			fmt.Printf("   Target: %s\n", host)
			fmt.Printf("   Packets: %d per hop\n", count)
			fmt.Println()
			fmt.Println("   Analisando rota... (pode demorar alguns segundos)")
			fmt.Println()

			result, err := usecase.RunMTR(ctx, host, count)
			if err != nil {
				return err
			}

			displayMTRResult(result)
			return nil
		},
	}

	cmd.Flags().IntVarP(&count, "count", "c", 10, "Número de pacotes por hop")

	return cmd
}

func displayMTRResult(result *models.MTRResult) {
	fmt.Println(strings.Repeat("─", 85))
	fmt.Printf(" %-4s %-40s %6s %8s %8s %8s %8s\n",
		"HOP", "HOST", "LOSS%", "AVG", "MIN", "MAX", "STDEV")
	fmt.Println(strings.Repeat("─", 85))

	for _, hop := range result.Hops {
		// Determine color based on status
		var rowStyle lipgloss.Style
		switch hop.Status {
		case models.HopStatusTimeout:
			rowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
		case models.HopStatusLoss:
			rowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		case models.HopStatusHighLat:
			rowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
		default:
			rowStyle = lipgloss.NewStyle()
		}

		// Format host (truncate if too long)
		hostDisplay := hop.Host
		if len(hostDisplay) > 38 {
			hostDisplay = hostDisplay[:35] + "..."
		}

		// Loss indicator
		lossStr := fmt.Sprintf("%.1f%%", hop.Loss)
		if hop.Loss > 0 {
			lossStr = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(lossStr)
		}

		// Latency display
		avgStr := "-"
		minStr := "-"
		maxStr := "-"
		stdStr := "-"

		if hop.AvgLatency > 0 {
			avgStr = fmt.Sprintf("%dms", hop.AvgLatency.Milliseconds())
			minStr = fmt.Sprintf("%dms", hop.MinLatency.Milliseconds())
			maxStr = fmt.Sprintf("%dms", hop.MaxLatency.Milliseconds())
			if hop.StdDev > 0 {
				stdStr = fmt.Sprintf("%dms", hop.StdDev.Milliseconds())
			}
		}

		fmt.Printf(" %s\n",
			rowStyle.Render(fmt.Sprintf("%-4d %-40s %6s %8s %8s %8s %8s",
				hop.Number,
				hostDisplay,
				lossStr,
				avgStr,
				minStr,
				maxStr,
				stdStr)))
	}

	fmt.Println(strings.Repeat("─", 85))

	// Summary
	fmt.Println()
	fmt.Println(" ANÁLISE")
	fmt.Println(strings.Repeat("─", 85))

	if result.Completed {
		fmt.Printf("   %s Destino alcançado em %d hops\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓"),
			result.TotalHops)
	} else {
		fmt.Printf("   %s Destino NÃO alcançado\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("✗"))
	}

	fmt.Printf("   Latência total: ~%s\n", result.Summary.TotalLatency.Round(time.Millisecond))

	if result.Summary.Bottleneck != "" {
		fmt.Printf("\n   %s Gargalo detectado:\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render("⚠"))
		fmt.Printf("     %s\n", result.Summary.Bottleneck)
	}

	fmt.Printf("\n   💡 %s\n", result.Summary.Recommendation)
	fmt.Printf("\n   Tempo de análise: %s\n", result.Duration.Round(time.Millisecond))
	fmt.Println()
}
