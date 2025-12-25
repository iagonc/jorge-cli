package commands

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/timex"
	"github.com/spf13/cobra"
)

// NewTimeCommand creates the time command
func NewTimeCommand(usecase *timex.TimeUsecase) *cobra.Command {
	var timezones []string

	cmd := &cobra.Command{
		Use:   "time [timestamp]",
		Short: "Converte timestamps e timezones",
		Long: `Utilitário para conversão de timestamps e timezones.

Se nenhum timestamp for fornecido, mostra o horário atual.

Formatos aceitos:
  - Unix timestamp (segundos, milissegundos, nanosegundos)
  - ISO8601 / RFC3339
  - "2006-01-02 15:04:05"
  - "02/01/2006 15:04:05" (formato BR)`,
		Example: `  # Horário atual
  jorge time
  jorge time --tz UTC,America/New_York

  # Converter timestamp
  jorge time 1703444400
  jorge time 1703444400000
  jorge time "2023-12-25T10:00:00Z"
  jorge time "25/12/2023 10:00:00"

  # Converter para múltiplos fusos
  jorge time 1703444400 --tz America/Sao_Paulo,Europe/London,Asia/Tokyo`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				// Show current time
				result := usecase.Now(timezones)
				displayTimeNow(result)
			} else {
				// Convert timestamp
				input := strings.Join(args, " ")
				result := usecase.Convert(input, timezones)
				displayTimeConversion(result)
			}
			return nil
		},
	}

	cmd.Flags().StringSliceVar(&timezones, "tz", nil, "Timezones para conversão (separados por vírgula)")

	// Add subcommands
	diffCmd := &cobra.Command{
		Use:   "diff <time1> <time2>",
		Short: "Calcula diferença entre dois timestamps",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			diff, err := usecase.Diff(args[0], args[1])
			if err != nil {
				return err
			}
			fmt.Println()
			fmt.Printf("   Diferença: %s\n\n",
				lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42")).Render(diff))
			return nil
		},
	}

	addCmd := &cobra.Command{
		Use:   "add <timestamp> <duration>",
		Short: "Adiciona duração a um timestamp",
		Long: `Adiciona uma duração a um timestamp.

Durações válidas: 1h, 30m, 24h, 7d (use 168h para 7 dias)`,
		Example: `  jorge time add 1703444400 24h
  jorge time add "2023-12-25" 168h`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := usecase.Add(args[0], args[1])
			if err != nil {
				return err
			}
			displayTimeConversion(result)
			return nil
		},
	}

	zonesCmd := &cobra.Command{
		Use:   "zones",
		Short: "Lista timezones comuns",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println()
			fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("🌍 Timezones Comuns"))
			fmt.Println()
			for _, tz := range usecase.CommonTimezones() {
				fmt.Printf("   • %s\n", tz)
			}
			fmt.Println()
		},
	}

	cmd.AddCommand(diffCmd, addCmd, zonesCmd)

	return cmd
}

func displayTimeNow(result *models.TimeNow) {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("🕐 Horário Atual"))
	fmt.Println()

	fmt.Printf("   Unix:        %s\n",
		lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("%d", result.Unix)))
	fmt.Printf("   Unix Milli:  %d\n", result.UnixMilli)
	fmt.Printf("   ISO8601:     %s\n", result.ISO8601)
	fmt.Printf("   RFC1123:     %s\n", result.RFC1123)
	fmt.Printf("   Human:       %s\n",
		lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(result.Human))

	if len(result.Timezones) > 0 {
		fmt.Println()
		fmt.Println(strings.Repeat("─", 50))
		fmt.Println("   TIMEZONES")
		fmt.Println(strings.Repeat("─", 50))
		for _, tz := range result.Timezones {
			fmt.Printf("   %-25s %s %s\n",
				tz.Timezone,
				lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(tz.Time),
				lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(tz.Offset))
		}
	}

	fmt.Println()
}

func displayTimeConversion(result *models.TimeConversion) {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("🕐 Conversão de Tempo"))
	fmt.Println()

	fmt.Printf("   Input:       %s\n", result.Input)
	fmt.Printf("   Formato:     %s\n",
		lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(result.InputFormat))

	if result.InputFormat == "Unknown" {
		fmt.Printf("\n   %s Não foi possível parsear o input\n\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("✗"))
		return
	}

	fmt.Println()
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println("   FORMATOS")
	fmt.Println(strings.Repeat("─", 50))

	fmt.Printf("   Unix:        %s\n",
		lipgloss.NewStyle().Bold(true).Render(result.Outputs["unix"]))
	fmt.Printf("   Unix Milli:  %s\n", result.Outputs["unix_milli"])
	fmt.Printf("   ISO8601:     %s\n", result.Outputs["iso8601"])
	fmt.Printf("   Date:        %s\n", result.Outputs["date"])
	fmt.Printf("   Time:        %s\n", result.Outputs["time"])
	fmt.Printf("   Human:       %s\n",
		lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(result.Outputs["human"]))
	fmt.Printf("   Relativo:    %s\n",
		lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render(result.Outputs["human_relative"]))

	if len(result.Timezones) > 0 {
		fmt.Println()
		fmt.Println(strings.Repeat("─", 50))
		fmt.Println("   TIMEZONES")
		fmt.Println(strings.Repeat("─", 50))
		for _, tz := range result.Timezones {
			fmt.Printf("   %-25s %s %s\n",
				tz.Timezone,
				lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(tz.Time),
				lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(tz.Offset))
		}
	}

	fmt.Println()
}
