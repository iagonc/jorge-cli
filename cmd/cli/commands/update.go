package commands

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/resource"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/utils"

	"go.uber.org/zap"
)

func NewUpdateCommand(usecase *resource.ResourceUsecase) *cobra.Command {
    var id, name, dns string

    cmd := &cobra.Command{
        Use:   "update",
        Short: "Update an existing resource",
        Run: func(cmd *cobra.Command, args []string) {
            ctx := cmd.Context()

            idInt, err := utils.ParseID(id)
            if err != nil {
                usecase.Logger.Error("Invalid ID", zap.Error(err))
                fmt.Println(err)
                return
            }

            if name == "" && dns == "" {
                fmt.Println("At least one of 'name' or 'dns' must be provided")
                return
            }

            updatedResource, err := usecase.UpdateResource(ctx, idInt, name, dns)
            if err != nil {
                usecase.Logger.Error("Error updating resource", zap.Error(err))
                fmt.Println("Error updating resource:", err)
                return
            }

            successStyle := lipgloss.NewStyle().
                Bold(true).
                Foreground(lipgloss.Color("#FFD700")). // Gold color
                Padding(1, 2).
                Align(lipgloss.Center)

            result := successStyle.Render(
                fmt.Sprintf("Resource Updated:\nID: %d\nName: %s\nDNS: %s",
                    updatedResource.ID, updatedResource.Name, updatedResource.Dns),
            )

            fmt.Println(result)
        },
    }

    cmd.Flags().StringVarP(&id, "id", "i", "", "Resource ID (required)")
    cmd.Flags().StringVarP(&name, "name", "n", "", "New resource name")
    cmd.Flags().StringVarP(&dns, "dns", "d", "", "New resource DNS")
    cmd.MarkFlagRequired("id")

    return cmd
}
