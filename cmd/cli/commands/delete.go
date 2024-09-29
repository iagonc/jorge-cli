package commands

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/resource"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/utils"

	"go.uber.org/zap"
)

func NewDeleteCommand(usecase *resource.ResourceUsecase) *cobra.Command {
	var id string

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a resource by ID",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()

			idInt, err := utils.ParseID(id)
			if err != nil {
				usecase.Logger.Error("Invalid ID", zap.Error(err))
				fmt.Println(err)
				return
			}

			resource, err := usecase.GetResourceByID(ctx, idInt)
			if err != nil {
				usecase.Logger.Error("Error fetching resource", zap.Error(err))
				fmt.Println("Error fetching resource:", err)
				return
			}

			fmt.Printf("Resource Details:\nID: %d\nName: %s\nDNS: %s\n", resource.ID, resource.Name, resource.Dns)

			if !utils.ConfirmAction("Are you sure you want to delete this resource? (yes/no): ") {
				fmt.Println("Delete operation canceled.")
				return
			}

			deletedResource, err := usecase.DeleteResource(ctx, idInt)
			if err != nil {
				usecase.Logger.Error("Error deleting resource", zap.Error(err))
				fmt.Println("Error deleting resource:", err)
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

	cmd.Flags().StringVarP(&id, "id", "i", "", "Resource ID (required)")
	cmd.MarkFlagRequired("id")

	return cmd
}
