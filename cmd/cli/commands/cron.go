package commands

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/cronx"
	"github.com/spf13/cobra"
)

// NewCronCommand creates the cron command
func NewCronCommand(usecase *cronx.CronUsecase) *cobra.Command {
	var nextCount int

	cmd := &cobra.Command{
		Use:   "cron <expression>",
		Short: "Interpreta expressões cron",
		Long: `Interpreta expressões cron e mostra próximas execuções.

Formato cron (5 campos):
  ┌───────────── minuto (0-59)
  │ ┌───────────── hora (0-23)
  │ │ ┌───────────── dia do mês (1-31)
  │ │ │ ┌───────────── mês (1-12)
  │ │ │ │ ┌───────────── dia da semana (0-6, Domingo=0)
  │ │ │ │ │
  * * * * *

Caracteres especiais:
  *     Qualquer valor
  */n   A cada n unidades
  n-m   Range de n a m
  n,m   Lista de valores`,
		Example: `  # Interpretar expressão
  jorge cron "0 * * * *"           # A cada hora
  jorge cron "*/15 * * * *"        # A cada 15 minutos
  jorge cron "0 0 * * *"           # Toda meia-noite
  jorge cron "0 9 * * 1-5"         # 9h de seg a sex
  jorge cron "0 0 1 * *"           # 1º de cada mês

  # Ver próximas N execuções
  jorge cron "*/30 * * * *" --next 10`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			expression := args[0]

			result := usecase.NextRuns(expression, nextCount)

			displayCronResult(result, nextCount)
			return nil
		},
	}

	cmd.Flags().IntVarP(&nextCount, "next", "n", 5, "Número de próximas execuções a mostrar")

	return cmd
}

func displayCronResult(result *models.CronExpression, count int) {
	fmt.Println()

	if !result.IsValid {
		fmt.Printf("   %s %s\n\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("✗"),
			result.Error)
		return
	}

	// Expression
	fmt.Printf("   %s\n",
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("⏰ Expressão Cron"))
	fmt.Println()

	fmt.Printf("   Expressão:   %s\n",
		lipgloss.NewStyle().Bold(true).Render(result.Expression))

	fmt.Printf("   Descrição:   %s\n",
		lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(result.Description))

	// Fields breakdown
	if result.Fields != nil {
		fmt.Println()
		fmt.Println(strings.Repeat("─", 50))
		fmt.Println("   CAMPOS")
		fmt.Println(strings.Repeat("─", 50))
		fmt.Printf("   Minuto:      %s\n", result.Fields.Minute)
		fmt.Printf("   Hora:        %s\n", result.Fields.Hour)
		fmt.Printf("   Dia do Mês:  %s\n", result.Fields.DayOfMonth)
		fmt.Printf("   Mês:         %s\n", result.Fields.Month)
		fmt.Printf("   Dia Semana:  %s\n", result.Fields.DayOfWeek)
	}

	// Next runs
	if len(result.NextRuns) > 0 {
		fmt.Println()
		fmt.Println(strings.Repeat("─", 50))
		fmt.Printf("   PRÓXIMAS %d EXECUÇÕES\n", len(result.NextRuns))
		fmt.Println(strings.Repeat("─", 50))

		for i, t := range result.NextRuns {
			weekday := []string{"Dom", "Seg", "Ter", "Qua", "Qui", "Sex", "Sáb"}[t.Weekday()]
			fmt.Printf("   %d. %s  %s\n",
				i+1,
				lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render(t.Format("02/01/2006 15:04")),
				lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(weekday))
		}
	}

	fmt.Println()
}
