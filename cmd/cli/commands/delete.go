package commands

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/iagonc/jorge-cli/cmd/cli/pkg/services"
	"github.com/iagonc/jorge-cli/cmd/cli/pkg/utils"

	"go.uber.org/zap"
)

func NewDeleteCommand(service *services.ResourceService) *cobra.Command {
    var id string

    cmd := &cobra.Command{
        Use:   "delete",
        Short: "Delete a resource by ID",
        Run: func(cmd *cobra.Command, args []string) {
            ctx := cmd.Context()
            idInt, err := strconv.Atoi(id)
            if err != nil {
                service.Logger.Error("--id must be a valid integer")
                return
            }

            resource, err := service.GetResourceByID(ctx, idInt)
            if err != nil {
                service.Logger.Error("Error fetching resource", zap.Error(err))
                return
            }

            // Display resource details
            fmt.Printf("Resource Details:\nID: %d\nName: %s\nDNS: %s\n", resource.ID, resource.Name, resource.Dns)

            // Ask for confirmation
            if !utils.ConfirmAction("Are you sure you want to delete this resource? (yes/no): ") {
                fmt.Println("Delete operation canceled.")
                return
            }

            // Proceed with deletion
            deletedResource, err := service.DeleteResource(ctx, idInt)
            if err != nil {
                service.Logger.Error("Error deleting resource", zap.Error(err))
                return
            }

            successStyle := lipgloss.NewStyle().
                Foreground(lipgloss.Color("#FF6347")). // Soft red color
                Bold(true)

            result := successStyle.Render(
                fmt.Sprintf("Resource Deleted:\nID: %d\nName: %s\nDNS: %s",
                    deletedResource.ID, deletedResource.Name, deletedResource.Dns),
            )

            fmt.Println(result)
        },
    }

    // Add flag for "id"
    cmd.Flags().StringVarP(&id, "id", "i", "", "Resource ID (required)")
    cmd.MarkFlagRequired("id")

    return cmd
}
