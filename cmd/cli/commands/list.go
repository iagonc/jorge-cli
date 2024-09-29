package commands

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/iagonc/jorge-cli/cmd/cli/pkg/services"

	"go.uber.org/zap"
)

func NewListCommand(service *services.ResourceService) *cobra.Command {
    return &cobra.Command{
        Use:   "list",
        Short: "Lista todos os recursos",
        Run: func(cmd *cobra.Command, args []string) {
            ctx := cmd.Context()
            resources, err := service.ListResources(ctx)
            if err != nil {
                service.Logger.Error("Erro ao listar recursos", zap.Error(err))
                return
            }

            // Estilos com Lipgloss
            headerStyle := lipgloss.NewStyle().
                Bold(true).
                Foreground(lipgloss.Color("#FAFAFA")).
                Background(lipgloss.Color("#7D56F4")).
                Padding(0, 1).
                Align(lipgloss.Left)

            rowStyle := lipgloss.NewStyle().
                Padding(0, 1).
                BorderStyle(lipgloss.NormalBorder()).
                BorderForeground(lipgloss.Color("#7D56F4"))

            tableHeader := headerStyle.Render(fmt.Sprintf("%-5s %-20s %-30s %-20s %-20s", "ID", "Name", "DNS", "CreatedAt", "UpdatedAt"))
            fmt.Println(tableHeader)

            for _, resource := range resources {
                createdAtFormatted := formatDate(resource.CreatedAt)
                updatedAtFormatted := formatDate(resource.UpdatedAt)

                resourceRow := fmt.Sprintf(
                    "%-5d %-20s %-30s %-20s %-20s",
                    resource.ID, resource.Name, resource.Dns, createdAtFormatted, updatedAtFormatted,
                )
                fmt.Println(rowStyle.Render(resourceRow))
            }
        },
    }
}

// formatDate formata a string de data para um formato legível
func formatDate(dateStr string) string {
    parsedTime, err := time.Parse(time.RFC3339Nano, dateStr)
    if err != nil {
        return dateStr
    }
    return parsedTime.Format("2006-01-02 15:04")
}
