package commands

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/iagonc/jorge-cli/cmd/cli/pkg/services"
	"github.com/iagonc/jorge-cli/cmd/cli/pkg/utils"

	"go.uber.org/zap"
)

func NewUpdateCommand(service *services.ResourceService) *cobra.Command {
    var id, name, dns string

    cmd := &cobra.Command{
        Use:   "update",
        Short: "Update an existing resource",
        Run: func(cmd *cobra.Command, args []string) {
            ctx := cmd.Context()

            idInt, err := utils.ParseID(id)
            if err != nil {
                service.Logger.Error("Invalid ID", zap.Error(err))
                fmt.Println(err)
                return
            }

            if name == "" && dns == "" {
                fmt.Println("At least one of 'name' or 'dns' must be provided")
                return
            }

            updatedResource, err := service.UpdateResource(ctx, idInt, name, dns)
            if err != nil {
                service.Logger.Error("Error updating resource", zap.Error(err))
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

    // Add flags for "id", "name", and "dns"
    cmd.Flags().StringVarP(&id, "id", "i", "", "Resource ID (required)")
    cmd.Flags().StringVarP(&name, "name", "n", "", "New resource name")
    cmd.Flags().StringVarP(&dns, "dns", "d", "", "New resource DNS")
    cmd.MarkFlagRequired("id")

    return cmd
}
