package commands

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/resource"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/utils"

	"go.uber.org/zap"
)

func NewListCommand(usecase *resource.ResourceUsecase) *cobra.Command {
    return &cobra.Command{
        Use:   "list",
        Short: "List all resources",
        Run: func(cmd *cobra.Command, args []string) {
            ctx := cmd.Context()
            resources, err := usecase.ListResources(ctx)
            if err != nil {
                usecase.Logger.Error("Error listing resources", zap.Error(err))
                fmt.Println("Error listing resources:", err)
                return
            }

            // Styles with Lipgloss
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
                createdAtFormatted := utils.FormatDate(resource.CreatedAt)
                updatedAtFormatted := utils.FormatDate(resource.UpdatedAt)

                resourceRow := fmt.Sprintf(
                    "%-5d %-20s %-30s %-20s %-20s",
                    resource.ID, resource.Name, resource.Dns, createdAtFormatted, updatedAtFormatted,
                )
                fmt.Println(rowStyle.Render(resourceRow))
            }
        },
    }
}
