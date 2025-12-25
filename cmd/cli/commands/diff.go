package commands

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/diff"
	"github.com/spf13/cobra"
)

// NewDiffCommand creates the diff command
func NewDiffCommand(usecase *diff.DiffUsecase) *cobra.Command {
	var semantic bool

	cmd := &cobra.Command{
		Use:   "diff <file1> <file2>",
		Short: "Compara dois arquivos (texto ou estruturado)",
		Long: `Compara dois arquivos e mostra as diferenças.

Modos:
  - Texto: Comparação linha a linha
  - Semântico: Comparação estruturada de JSON/YAML

O modo semântico entende a estrutura do arquivo e mostra
diferenças por chave/valor ao invés de por linha.`,
		Example: `  # Diff de texto
  jorge diff config-prod.yaml config-staging.yaml

  # Diff semântico (JSON/YAML)
  jorge diff config-prod.yaml config-staging.yaml --semantic

  # Diff de arquivos .env
  jorge diff .env.prod .env.staging`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := usecase.DiffFiles(args[0], args[1], semantic)
			if err != nil {
				return err
			}

			displayDiffResult(result)
			return nil
		},
	}

	cmd.Flags().BoolVar(&semantic, "semantic", false, "Usar diff semântico (para JSON/YAML)")

	// Add env subcommand
	envCmd := &cobra.Command{
		Use:   "env <file1> <file2>",
		Short: "Compara dois arquivos .env",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := usecase.DiffEnv(args[0], args[1])
			if err != nil {
				return err
			}

			displayDiffResult(result)
			return nil
		},
	}

	cmd.AddCommand(envCmd)

	return cmd
}

func displayDiffResult(result *models.DiffResult) {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("📋 Diff"))
	fmt.Println()

	fmt.Printf("   %s vs %s\n",
		lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(result.File1),
		lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(result.File2))
	fmt.Println()

	if result.Identical {
		fmt.Printf("   %s Arquivos idênticos\n\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓"))
		return
	}

	// Stats
	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("   Adicionados: %s  Removidos: %s  Modificados: %s\n",
		lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true).Render(fmt.Sprintf("+%d", result.Stats.Added)),
		lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).Render(fmt.Sprintf("-%d", result.Stats.Removed)),
		lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Render(fmt.Sprintf("~%d", result.Stats.Modified)))
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println()

	// Changes
	for _, change := range result.Changes {
		switch change.Type {
		case models.DiffTypeAdded:
			location := ""
			if change.Path != "" {
				location = change.Path
			} else if change.Line > 0 {
				location = fmt.Sprintf("linha %d", change.Line)
			}

			fmt.Printf("   %s %s\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("+"),
				lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(location))
			fmt.Printf("      %s\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(truncate(change.NewValue, 70)))

		case models.DiffTypeRemoved:
			location := ""
			if change.Path != "" {
				location = change.Path
			} else if change.Line > 0 {
				location = fmt.Sprintf("linha %d", change.Line)
			}

			fmt.Printf("   %s %s\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("-"),
				lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(location))
			fmt.Printf("      %s\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(truncate(change.OldValue, 70)))

		case models.DiffTypeModified:
			location := ""
			if change.Path != "" {
				location = change.Path
			} else if change.Line > 0 {
				location = fmt.Sprintf("linha %d", change.Line)
			}

			fmt.Printf("   %s %s\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render("~"),
				lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(location))
			fmt.Printf("      %s %s\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("-"),
				lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(truncate(change.OldValue, 65)))
			fmt.Printf("      %s %s\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("+"),
				lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(truncate(change.NewValue, 65)))
		}
		fmt.Println()
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
