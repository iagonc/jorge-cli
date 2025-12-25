package commands

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/secrets"
	"github.com/spf13/cobra"
)

// NewSecretsCommand creates the secrets command
func NewSecretsCommand(usecase *secrets.SecretsUsecase) *cobra.Command {
	var excludes []string

	cmd := &cobra.Command{
		Use:   "secrets <path>",
		Short: "Escaneia segredos expostos no codigo",
		Long: `Escaneia um diretorio ou arquivo em busca de segredos expostos.

Detecta:
  - Chaves AWS (Access Key ID e Secret)
  - Tokens GitHub (PAT e fine-grained)
  - Tokens Slack
  - Chaves privadas (RSA, EC, PGP)
  - API Keys e Secrets genericos
  - Tokens JWT
  - Senhas em arquivos de config
  - Chaves GCP (service account)`,
		Example: `  # Escanear diretorio atual
  jorge secrets .

  # Escanear diretorio especifico
  jorge secrets /path/to/project

  # Excluir arquivos
  jorge secrets . --exclude "*.test.js" --exclude "*.md"

  # Escanear arquivo unico
  jorge secrets config.env`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := usecase.Scan(args[0], excludes)
			if err != nil {
				return err
			}

			displaySecretsResult(result)
			return nil
		},
	}

	cmd.Flags().StringArrayVar(&excludes, "exclude", nil, "Patterns para excluir")

	// Subcommand to list patterns
	patternsCmd := &cobra.Command{
		Use:   "patterns",
		Short: "Lista os padroes de deteccao",
		Run: func(cmd *cobra.Command, args []string) {
			patterns := usecase.GetPatterns()
			displaySecretPatterns(patterns)
		},
	}

	cmd.AddCommand(patternsCmd)

	return cmd
}

func displaySecretsResult(result *models.SecretScanResult) {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("🔐 Secret Scan"))
	fmt.Println()

	// Stats
	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("   Path: %s\n", result.Path)
	fmt.Printf("   Arquivos escaneados: %d | Ignorados: %d\n",
		result.ScannedFiles, result.SkippedFiles)
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println()

	if result.TotalFindings == 0 {
		fmt.Printf("   %s Nenhum segredo encontrado!\n\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓"))
		return
	}

	// Group by severity
	highStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	mediumStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	lowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226"))

	var high, medium, low int
	for _, f := range result.Findings {
		switch f.Severity {
		case "high":
			high++
		case "medium":
			medium++
		default:
			low++
		}
	}

	fmt.Printf("   Encontrados: %s %s %s\n\n",
		highStyle.Render(fmt.Sprintf("%d high", high)),
		mediumStyle.Render(fmt.Sprintf("%d medium", medium)),
		lowStyle.Render(fmt.Sprintf("%d low", low)))

	// Findings
	for _, finding := range result.Findings {
		var severityBadge string
		switch finding.Severity {
		case "high":
			severityBadge = highStyle.Render("[HIGH]")
		case "medium":
			severityBadge = mediumStyle.Render("[MED]")
		default:
			severityBadge = lowStyle.Render("[LOW]")
		}

		fmt.Printf("   %s %s\n", severityBadge, finding.Description)
		fmt.Printf("      %s:%d\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(finding.File),
			finding.Line)
		fmt.Printf("      Match: %s\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Render(finding.Match))
		fmt.Println()
	}
}

func displaySecretPatterns(patterns []models.SecretPattern) {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("🔐 Secret Patterns"))
	fmt.Println()

	highStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	mediumStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))

	for _, p := range patterns {
		var sev string
		if p.Severity == "high" {
			sev = highStyle.Render("[HIGH]")
		} else {
			sev = mediumStyle.Render("[MED]")
		}

		fmt.Printf("   %s %s\n", sev, p.Description)
		fmt.Printf("      Type: %s\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(string(p.Type)))
		fmt.Println()
	}
}
