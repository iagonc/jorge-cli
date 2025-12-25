package commands

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/logs"
	"github.com/spf13/cobra"
)

// NewLogsCommand creates the logs command
func NewLogsCommand(usecase *logs.LogsUsecase) *cobra.Command {
	var format string
	var level string
	var pattern string
	var stats bool
	var last string
	var limit int

	cmd := &cobra.Command{
		Use:   "logs [file]",
		Short: "Analisa logs e extrai estatísticas",
		Long: `Analisa arquivos de log e extrai estatísticas úteis.

Formatos suportados:
  - json:   Logs em formato JSON
  - nginx:  Combined log format do Nginx
  - apache: Combined log format do Apache
  - syslog: Formato syslog padrão
  - common: Detecção automática

Funcionalidades:
  - Contagem por nível (ERROR, WARN, INFO, etc)
  - Taxa de erros
  - Top erros mais frequentes
  - Filtro por padrão ou nível
  - Filtro por tempo`,
		Example: `  # Analisar arquivo de log
  jorge logs app.log
  jorge logs /var/log/nginx/access.log --format nginx

  # Mostrar apenas estatísticas
  jorge logs app.log --stats

  # Filtrar por nível
  jorge logs app.log --level error

  # Filtrar por padrão
  jorge logs app.log --pattern "timeout|connection"

  # Últimas N horas
  jorge logs app.log --last 1h

  # De stdin
  kubectl logs pod-name | jorge logs --format json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var reader io.Reader

			if len(args) > 0 {
				file, err := os.Open(args[0])
				if err != nil {
					return err
				}
				defer file.Close()
				reader = file
			} else {
				reader = os.Stdin
			}

			logFormat := models.LogFormat(format)
			if format == "" || format == "auto" {
				logFormat = models.LogFormatCommon
			}

			logStats, entries, err := usecase.Analyze(reader, logFormat)
			if err != nil {
				return err
			}

			// Apply filters
			var since time.Time
			if last != "" {
				duration, err := time.ParseDuration(last)
				if err == nil {
					since = time.Now().Add(-duration)
				}
			}

			if level != "" || pattern != "" || !since.IsZero() {
				entries = usecase.Filter(entries, level, pattern, since, time.Time{})
			}

			if stats {
				displayLogStats(logStats)
			} else {
				displayLogEntries(entries, logStats, limit)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "auto", "Formato do log (json, nginx, apache, syslog, auto)")
	cmd.Flags().StringVarP(&level, "level", "l", "", "Filtrar por nível (error, warn, info, debug)")
	cmd.Flags().StringVarP(&pattern, "pattern", "p", "", "Filtrar por regex")
	cmd.Flags().BoolVar(&stats, "stats", false, "Mostrar apenas estatísticas")
	cmd.Flags().StringVar(&last, "last", "", "Mostrar apenas últimas N (ex: 1h, 30m)")
	cmd.Flags().IntVarP(&limit, "limit", "n", 100, "Limite de linhas a mostrar")

	return cmd
}

func displayLogStats(stats *models.LogStats) {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("📊 Estatísticas de Log"))
	fmt.Println()

	fmt.Println(strings.Repeat("─", 50))
	fmt.Println("   RESUMO")
	fmt.Println(strings.Repeat("─", 50))
	fmt.Printf("   Total de linhas:    %d\n", stats.TotalLines)
	fmt.Printf("   Linhas parseadas:   %d\n", stats.ParsedLines)
	fmt.Printf("   Linhas com falha:   %d\n", stats.FailedLines)

	if stats.TimeRange != nil {
		fmt.Printf("   Período:            %s - %s\n",
			stats.TimeRange.Start.Format("02/01 15:04"),
			stats.TimeRange.End.Format("02/01 15:04"))
	}

	fmt.Println()
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println("   NÍVEIS")
	fmt.Println(strings.Repeat("─", 50))

	levelColors := map[string]string{
		"ERROR": "196", "FATAL": "196", "CRITICAL": "196",
		"WARN": "214", "WARNING": "214",
		"INFO": "42",
		"DEBUG": "245", "TRACE": "245",
	}

	for level, count := range stats.LevelCounts {
		color := levelColors[level]
		if color == "" {
			color = "255"
		}
		bar := strings.Repeat("█", min(count*50/max(stats.ParsedLines, 1), 50))
		fmt.Printf("   %-10s %5d  %s\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(level),
			count,
			lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(bar))
	}

	// Error rate
	fmt.Println()
	rateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	if stats.ErrorRate > 5 {
		rateStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	} else if stats.ErrorRate > 1 {
		rateStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	}
	fmt.Printf("   Taxa de erros: %s\n",
		rateStyle.Bold(true).Render(fmt.Sprintf("%.2f%%", stats.ErrorRate)))

	// Top errors
	if len(stats.TopErrors) > 0 {
		fmt.Println()
		fmt.Println(strings.Repeat("─", 50))
		fmt.Println("   TOP ERROS")
		fmt.Println(strings.Repeat("─", 50))

		for i, e := range stats.TopErrors {
			msg := e.Message
			if len(msg) > 60 {
				msg = msg[:57] + "..."
			}
			fmt.Printf("   %d. (%d) %s\n", i+1, e.Count,
				lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(msg))
		}
	}

	fmt.Println()
}

func displayLogEntries(entries []models.LogEntry, stats *models.LogStats, limit int) {
	fmt.Println()

	// Show summary first
	fmt.Printf("   Mostrando %d de %d entradas\n\n",
		min(len(entries), limit),
		len(entries))

	levelColors := map[string]string{
		"ERROR": "196", "FATAL": "196", "CRITICAL": "196",
		"WARN": "214", "WARNING": "214",
		"INFO": "42",
		"DEBUG": "245", "TRACE": "245",
	}

	for i, entry := range entries {
		if i >= limit {
			fmt.Printf("\n   ... e mais %d entradas\n", len(entries)-limit)
			break
		}

		// Format timestamp
		timeStr := ""
		if !entry.Timestamp.IsZero() {
			timeStr = entry.Timestamp.Format("15:04:05")
		}

		// Format level
		level := entry.Level
		if level == "" {
			level = "---"
		}
		color := levelColors[strings.ToUpper(level)]
		if color == "" {
			color = "245"
		}

		// Format message
		msg := entry.Message
		if msg == "" {
			msg = entry.Raw
		}
		if len(msg) > 100 {
			msg = msg[:97] + "..."
		}

		fmt.Printf("   %s %s %s\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(timeStr),
			lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Width(7).Render(level),
			msg)
	}

	fmt.Println()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
