package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/diagnose"
	"github.com/spf13/cobra"
)

// Styles for diagnose output
var (
	diagTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			MarginBottom(1)

	diagTargetStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true)

	diagPassedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)

	diagWarningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true)

	diagFailedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	diagCategoryStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	diagSuggestionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("147")).
			PaddingLeft(2)

	diagCommandStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("220")).
			Background(lipgloss.Color("236")).
			Padding(0, 1)
)

// NewDiagnoseCommand creates the diagnose command
func NewDiagnoseCommand(usecase *diagnose.DiagnoseUsecase) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diagnose <target>",
		Short: "Diagnóstico completo de conectividade",
		Long: `Realiza um diagnóstico completo de rede para um host ou URL.

Executa automaticamente:
  - Resolução DNS
  - Teste de conectividade TCP
  - Verificação de certificado SSL
  - Teste de resposta HTTP
  - Medição de latência
  - Verificação de portas comuns

O resultado inclui problemas encontrados e sugestões de correção.`,
		Example: `  # Diagnosticar um domínio
  jorge-cli diagnose google.com

  # Diagnosticar uma URL completa
  jorge-cli diagnose https://api.exemplo.com/health

  # Diagnosticar com porta específica
  jorge-cli diagnose meuservidor.com:8080`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]
			timeout, _ := cmd.Flags().GetInt("timeout")

			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
			defer cancel()

			fmt.Println()
			fmt.Println(diagTitleStyle.Render("🔍 Diagnóstico de Rede"))
			fmt.Printf("   Target: %s\n\n", diagTargetStyle.Render(target))

			// Show spinner-like message
			fmt.Println("   Executando verificações...")
			fmt.Println()

			result, err := usecase.Diagnose(ctx, target)
			if err != nil {
				return fmt.Errorf("erro no diagnóstico: %w", err)
			}

			displayDiagnoseResult(result)
			return nil
		},
	}

	cmd.Flags().IntP("timeout", "t", 30, "Timeout em segundos")

	return cmd
}

func displayDiagnoseResult(result *models.DiagnoseResult) {
	// Display checks
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println(" VERIFICAÇÕES")
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println()

	for _, check := range result.Checks {
		displayCheck(check)
	}

	// Display summary
	fmt.Println()
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println(" RESULTADO")
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println()

	// Overall status
	var statusIcon, statusText string
	var statusStyle lipgloss.Style
	switch result.Overall {
	case models.StatusHealthy:
		statusIcon = "✅"
		statusText = "TUDO OK"
		statusStyle = diagPassedStyle
	case models.StatusUnhealthy:
		statusIcon = "⚠️"
		statusText = "ATENÇÃO"
		statusStyle = diagWarningStyle
	case models.StatusError:
		statusIcon = "❌"
		statusText = "PROBLEMAS ENCONTRADOS"
		statusStyle = diagFailedStyle
	}

	fmt.Printf("   %s %s\n", statusIcon, statusStyle.Render(statusText))
	fmt.Printf("   %s\n", result.Summary)
	fmt.Printf("   Tempo total: %v\n", result.Duration.Round(time.Millisecond))

	// Display problems if any
	if len(result.Problems) > 0 {
		fmt.Println()
		fmt.Println(strings.Repeat("─", 60))
		fmt.Println(" PROBLEMAS ENCONTRADOS")
		fmt.Println(strings.Repeat("─", 60))
		fmt.Println()

		for i, problem := range result.Problems {
			var icon string
			var style lipgloss.Style
			switch problem.Severity {
			case models.SeverityCritical:
				icon = "🔴"
				style = diagFailedStyle
			case models.SeverityWarning:
				icon = "🟡"
				style = diagWarningStyle
			default:
				icon = "🔵"
				style = diagCategoryStyle
			}

			fmt.Printf("   %s %s\n", icon, style.Render(problem.Title))
			fmt.Printf("      %s\n", problem.Description)
			if i < len(result.Problems)-1 {
				fmt.Println()
			}
		}
	}

	// Display suggestions if any
	if len(result.Suggestions) > 0 {
		fmt.Println()
		fmt.Println(strings.Repeat("─", 60))
		fmt.Println(" 💡 SUGESTÕES")
		fmt.Println(strings.Repeat("─", 60))
		fmt.Println()

		for i, suggestion := range result.Suggestions {
			fmt.Printf("   %d. %s\n", i+1, lipgloss.NewStyle().Bold(true).Render(suggestion.Title))
			fmt.Printf("      %s\n", suggestion.Description)
			if suggestion.Command != "" {
				fmt.Printf("      Comando: %s\n", diagCommandStyle.Render(suggestion.Command))
			}
			if suggestion.Link != "" {
				fmt.Printf("      Link: %s\n", suggestion.Link)
			}
			if i < len(result.Suggestions)-1 {
				fmt.Println()
			}
		}
	}

	fmt.Println()
}

func displayCheck(check models.DiagnoseCheck) {
	var icon string
	var style lipgloss.Style

	switch check.Status {
	case models.CheckPassed:
		icon = "✓"
		style = diagPassedStyle
	case models.CheckWarning:
		icon = "!"
		style = diagWarningStyle
	case models.CheckFailed:
		icon = "✗"
		style = diagFailedStyle
	case models.CheckSkipped:
		icon = "-"
		style = diagCategoryStyle
	}

	category := diagCategoryStyle.Render(fmt.Sprintf("[%s]", check.Category))
	status := style.Render(fmt.Sprintf("[%s]", icon))

	fmt.Printf("   %s %s %s\n", status, category, check.Name)
	fmt.Printf("      %s\n", style.Render(check.Message))
	if check.Details != "" {
		fmt.Printf("      %s\n", diagCategoryStyle.Render(check.Details))
	}
	fmt.Println()
}
