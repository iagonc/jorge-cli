package commands

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/explain"
	"github.com/spf13/cobra"
)

// NewExplainCommand creates the explain command
func NewExplainCommand(usecase *explain.ExplainUsecase) *cobra.Command {
	var listErrors bool

	cmd := &cobra.Command{
		Use:   "explain <error>",
		Short: "Explica um erro de rede em português simples",
		Long: `Explica erros de rede em linguagem simples e sugere soluções.

Cola a mensagem de erro que você recebeu e o jorge vai explicar:
  - O que significa o erro
  - Causas comuns
  - Como resolver (com comandos prontos)

Útil quando você recebe um erro e não sabe o que fazer.`,
		Example: `  # Explicar um erro específico
  jorge explain "connection refused"
  jorge explain "no such host"
  jorge explain "certificate has expired"
  jorge explain "timeout"

  # Listar todos os erros conhecidos
  jorge explain --list`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if listErrors {
				displayKnownErrors(usecase)
				return nil
			}

			if len(args) == 0 {
				return fmt.Errorf("especifique um erro para explicar ou use --list")
			}

			errorText := args[0]

			fmt.Println()
			explanation := usecase.Explain(errorText)
			displayExplanation(explanation)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&listErrors, "list", "l", false, "Listar todos os erros conhecidos")

	return cmd
}

func displayExplanation(exp *models.ErrorExplanation) {
	// Title with category
	categoryStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Background(lipgloss.Color("39")).
		Foreground(lipgloss.Color("0"))

	fmt.Println(strings.Repeat("─", 70))
	fmt.Printf(" %s  %s\n",
		categoryStyle.Render(string(exp.Category)),
		lipgloss.NewStyle().Bold(true).Render(exp.Title))
	fmt.Println(strings.Repeat("─", 70))
	fmt.Println()

	// Description
	fmt.Printf("   %s\n", exp.Description)
	fmt.Println()

	// Common causes
	if len(exp.CommonCauses) > 0 {
		fmt.Println(lipgloss.NewStyle().Bold(true).Render("   📋 Causas Comuns:"))
		fmt.Println()
		for _, cause := range exp.CommonCauses {
			fmt.Printf("      • %s\n", cause)
		}
		fmt.Println()
	}

	// Solutions
	if len(exp.Solutions) > 0 {
		fmt.Println(strings.Repeat("─", 70))
		fmt.Println(lipgloss.NewStyle().Bold(true).Render(" 🔧 SOLUÇÕES"))
		fmt.Println(strings.Repeat("─", 70))
		fmt.Println()

		for i, sol := range exp.Solutions {
			priorityStyle := lipgloss.NewStyle().
				Padding(0, 1).
				Background(lipgloss.Color("214")).
				Foreground(lipgloss.Color("0"))

			fmt.Printf("   %s %s\n",
				priorityStyle.Render(fmt.Sprintf("%d", i+1)),
				lipgloss.NewStyle().Bold(true).Render(sol.Title))

			fmt.Printf("      %s\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(sol.Description))

			if sol.Command != "" {
				fmt.Printf("      %s %s\n",
					lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render("$"),
					lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(sol.Command))
			}
			fmt.Println()
		}
	}

	// Related errors
	if len(exp.RelatedErrors) > 0 {
		fmt.Println(strings.Repeat("─", 70))
		fmt.Printf(" 🔗 Erros Relacionados: %s\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(strings.Join(exp.RelatedErrors, ", ")))
	}

	// Examples
	if len(exp.Examples) > 0 {
		fmt.Println()
		fmt.Println(" 📝 Exemplos de mensagens:")
		for _, ex := range exp.Examples {
			fmt.Printf("    %s\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Italic(true).Render(ex))
		}
	}

	// Links
	if len(exp.Links) > 0 {
		fmt.Println()
		fmt.Println(" 📚 Mais informações:")
		for _, link := range exp.Links {
			fmt.Printf("    %s\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render(link))
		}
	}

	fmt.Println()
}

func displayKnownErrors(usecase *explain.ExplainUsecase) {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("📚 Erros Conhecidos"))
	fmt.Println()

	// Get categories
	categories := usecase.GetCategories()

	fmt.Println(strings.Repeat("─", 50))
	fmt.Println(" CATEGORIAS")
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println()

	for cat, count := range categories {
		fmt.Printf("   • %s: %d erros\n", cat, count)
	}
	fmt.Println()

	// List all errors
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println(" ERROS")
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println()

	errors := usecase.ListKnownErrors()
	for _, err := range errors {
		fmt.Printf("   • %s\n", err)
	}
	fmt.Println()

	fmt.Printf("   Use: jorge explain \"<erro>\" para ver detalhes\n")
	fmt.Println()
}
