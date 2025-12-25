package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/headers"
	"github.com/spf13/cobra"
)

// NewHeadersCommand creates the headers command
func NewHeadersCommand(usecase *headers.HeadersUsecase) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "headers <url>",
		Short: "Analisa headers de segurança",
		Long: `Analisa os headers de segurança HTTP de um site.

Verifica:
  - HSTS (Strict-Transport-Security)
  - CSP (Content-Security-Policy)
  - X-Content-Type-Options
  - X-Frame-Options
  - Referrer-Policy
  - CORS
  - E outros...

Gera uma nota de A+ a F baseada na configuração.`,
		Example: `  jorge headers https://exemplo.com
  jorge headers https://api.exemplo.com`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			url := args[0]

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			fmt.Println()
			fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("🔒 Security Headers Analysis"))
			fmt.Printf("   URL: %s\n", url)
			fmt.Println()

			result, err := usecase.Analyze(ctx, url)
			if err != nil {
				return err
			}

			displayHeadersResult(result)
			return nil
		},
	}

	return cmd
}

func displayHeadersResult(result *models.HeadersResult) {
	// Grade display
	gradeStyle := lipgloss.NewStyle().Bold(true).Padding(0, 2)
	switch {
	case result.Grade == "A+" || result.Grade == "A":
		gradeStyle = gradeStyle.Background(lipgloss.Color("42")).Foreground(lipgloss.Color("0"))
	case result.Grade == "B":
		gradeStyle = gradeStyle.Background(lipgloss.Color("42")).Foreground(lipgloss.Color("0"))
	case result.Grade == "C":
		gradeStyle = gradeStyle.Background(lipgloss.Color("214")).Foreground(lipgloss.Color("0"))
	case result.Grade == "D":
		gradeStyle = gradeStyle.Background(lipgloss.Color("208")).Foreground(lipgloss.Color("0"))
	default:
		gradeStyle = gradeStyle.Background(lipgloss.Color("196")).Foreground(lipgloss.Color("255"))
	}

	fmt.Printf("   Nota: %s  Score: %d/100\n",
		gradeStyle.Render(result.Grade),
		result.SecurityScore)
	fmt.Println()

	// Headers check results
	fmt.Println(strings.Repeat("─", 70))
	fmt.Println(" VERIFICAÇÕES")
	fmt.Println(strings.Repeat("─", 70))
	fmt.Println()

	for _, check := range result.Checks {
		var icon string
		var statusStyle lipgloss.Style

		switch check.Status {
		case models.HeaderStatusPass:
			icon = "✓"
			statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
		case models.HeaderStatusWarn:
			icon = "!"
			statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
		case models.HeaderStatusFail:
			icon = "✗"
			statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		case models.HeaderStatusInfo:
			icon = "i"
			statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
		}

		presentStr := ""
		if check.Present {
			presentStr = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("presente")
		} else {
			presentStr = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("ausente")
		}

		fmt.Printf("   %s %-25s %s\n",
			statusStyle.Render(fmt.Sprintf("[%s]", icon)),
			check.Name,
			presentStr)

		fmt.Printf("      %s\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(check.Description))

		if check.Value != "" {
			// Truncate long values
			value := check.Value
			if len(value) > 60 {
				value = value[:57] + "..."
			}
			fmt.Printf("      Valor: %s\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render(value))
		}

		fmt.Println()
	}

	// Suggestions
	if len(result.Suggestions) > 0 {
		fmt.Println(strings.Repeat("─", 70))
		fmt.Println(" 💡 SUGESTÕES DE MELHORIA")
		fmt.Println(strings.Repeat("─", 70))
		fmt.Println()

		for i, suggestion := range result.Suggestions {
			fmt.Printf("   %d. %s\n", i+1, suggestion)
		}
		fmt.Println()
	}

	// Footer
	fmt.Println(strings.Repeat("─", 70))
	fmt.Printf("   HTTP Status: %d | Tempo: %s\n", result.StatusCode, result.Duration.Round(time.Millisecond))
	fmt.Println()
}
