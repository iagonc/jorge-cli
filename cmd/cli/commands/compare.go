package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/compare"
	"github.com/spf13/cobra"
)

// NewCompareCommand creates the compare command
func NewCompareCommand(usecase *compare.CompareUsecase) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compare <endpoint1> <endpoint2>",
		Short: "Compara dois endpoints",
		Long: `Compara conectividade e resposta entre dois endpoints.

Útil para:
  - Verificar diferenças entre ambientes (dev vs prod)
  - Comparar performance entre servidores
  - Identificar problemas de configuração`,
		Example: `  jorge compare api.dev.exemplo.com api.prod.exemplo.com
  jorge compare https://old.api.com https://new.api.com
  jorge compare server1:8080 server2:8080`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			endpoint1 := args[0]
			endpoint2 := args[1]

			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			fmt.Println()
			fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("🔀 Compare Endpoints"))
			fmt.Println()

			result, err := usecase.Compare(ctx, endpoint1, endpoint2)
			if err != nil {
				return err
			}

			displayCompareResult(result)
			return nil
		},
	}

	return cmd
}

func displayCompareResult(result *models.CompareResult) {
	// Header
	fmt.Println(strings.Repeat("─", 70))
	fmt.Printf("   %-30s vs %-30s\n", result.Endpoint1.Target, result.Endpoint2.Target)
	fmt.Println(strings.Repeat("─", 70))
	fmt.Println()

	// Side by side comparison
	displayCompareRow("Acessível",
		formatBool(result.Endpoint1.Reachable),
		formatBool(result.Endpoint2.Reachable),
		result.Endpoint1.Reachable == result.Endpoint2.Reachable)

	displayCompareRow("IP",
		result.Endpoint1.IP,
		result.Endpoint2.IP,
		true) // IPs can differ

	displayCompareRow("DNS Time",
		result.Endpoint1.DNSTime.Round(time.Millisecond).String(),
		result.Endpoint2.DNSTime.Round(time.Millisecond).String(),
		true)

	displayCompareRow("Connect Time",
		result.Endpoint1.ConnectTime.Round(time.Millisecond).String(),
		result.Endpoint2.ConnectTime.Round(time.Millisecond).String(),
		true)

	if result.Endpoint1.TLSTime > 0 || result.Endpoint2.TLSTime > 0 {
		displayCompareRow("TLS Time",
			result.Endpoint1.TLSTime.Round(time.Millisecond).String(),
			result.Endpoint2.TLSTime.Round(time.Millisecond).String(),
			true)

		displayCompareRow("TLS Version",
			result.Endpoint1.TLSVersion,
			result.Endpoint2.TLSVersion,
			result.Endpoint1.TLSVersion == result.Endpoint2.TLSVersion)
	}

	if result.Endpoint1.StatusCode > 0 || result.Endpoint2.StatusCode > 0 {
		displayCompareRow("HTTP Status",
			fmt.Sprintf("%d", result.Endpoint1.StatusCode),
			fmt.Sprintf("%d", result.Endpoint2.StatusCode),
			result.Endpoint1.StatusCode == result.Endpoint2.StatusCode)

		displayCompareRow("Response Time",
			result.Endpoint1.ResponseTime.Round(time.Millisecond).String(),
			result.Endpoint2.ResponseTime.Round(time.Millisecond).String(),
			true)
	}

	// Errors
	if result.Endpoint1.Error != "" {
		fmt.Printf("\n   %s Erro em %s: %s\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("✗"),
			result.Endpoint1.Target,
			result.Endpoint1.Error)
	}
	if result.Endpoint2.Error != "" {
		fmt.Printf("\n   %s Erro em %s: %s\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("✗"),
			result.Endpoint2.Target,
			result.Endpoint2.Error)
	}

	// Differences
	if len(result.Differences) > 0 {
		fmt.Println()
		fmt.Println(strings.Repeat("─", 70))
		fmt.Println(" DIFERENÇAS")
		fmt.Println(strings.Repeat("─", 70))

		for _, diff := range result.Differences {
			var icon string
			var color lipgloss.Color
			switch diff.Severity {
			case "critical":
				icon = "🔴"
				color = lipgloss.Color("196")
			case "warning":
				icon = "🟡"
				color = lipgloss.Color("214")
			default:
				icon = "🔵"
				color = lipgloss.Color("39")
			}

			fmt.Printf("\n   %s %s\n", icon, lipgloss.NewStyle().Foreground(color).Bold(true).Render(diff.Field))
			fmt.Printf("      %s: %s\n", result.Endpoint1.Target, diff.Value1)
			fmt.Printf("      %s: %s\n", result.Endpoint2.Target, diff.Value2)
			fmt.Printf("      %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(diff.Description))
		}
	}

	// Summary
	fmt.Println()
	fmt.Println(strings.Repeat("─", 70))

	if len(result.Differences) == 0 {
		fmt.Printf("   %s %s\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓"),
			result.Summary)
	} else {
		fmt.Printf("   %s\n", result.Summary)
	}

	fmt.Printf("   Tempo total: %s\n", result.Duration.Round(time.Millisecond))
	fmt.Println()
}

func displayCompareRow(label, value1, value2 string, match bool) {
	labelStyle := lipgloss.NewStyle().Width(15)
	valueStyle := lipgloss.NewStyle().Width(25)

	matchIcon := " "
	if !match {
		matchIcon = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("≠")
	}

	fmt.Printf("   %s %s %s %s\n",
		labelStyle.Render(label+":"),
		valueStyle.Render(value1),
		matchIcon,
		valueStyle.Render(value2))
}

func formatBool(b bool) string {
	if b {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("Sim")
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("Não")
}
