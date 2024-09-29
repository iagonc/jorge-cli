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
        Short: "Deleta um recurso pelo ID",
        Run: func(cmd *cobra.Command, args []string) {
            ctx := cmd.Context()
            idInt, err := strconv.Atoi(id)
            if err != nil {
                service.Logger.Error("--id deve ser um inteiro válido")
                return
            }

            resource, err := service.GetResourceByID(ctx, idInt)
            if err != nil {
                service.Logger.Error("Erro ao buscar recurso", zap.Error(err))
                return
            }

            // Exibe detalhes do recurso
            fmt.Printf("Detalhes do Recurso:\nID: %d\nName: %s\nDNS: %s\n", resource.ID, resource.Name, resource.Dns)

            // Solicita confirmação
            if !utils.ConfirmAction("Tem certeza de que deseja deletar este recurso? (yes/no): ") {
                fmt.Println("Operação de exclusão cancelada.")
                return
            }

            // Prossegue com a exclusão
            deletedResource, err := service.DeleteResource(ctx, idInt)
            if err != nil {
                service.Logger.Error("Erro ao deletar recurso", zap.Error(err))
                return
            }

            successStyle := lipgloss.NewStyle().
                Foreground(lipgloss.Color("#FF6347")). // Cor vermelho suave
                Bold(true)

            result := successStyle.Render(
                fmt.Sprintf("Recurso Deletado:\nID: %d\nName: %s\nDNS: %s",
                    deletedResource.ID, deletedResource.Name, deletedResource.Dns),
            )

            fmt.Println(result)
        },
    }

    // Adiciona flag para "id"
    cmd.Flags().StringVarP(&id, "id", "i", "", "ID do recurso (obrigatório)")
    cmd.MarkFlagRequired("id")

    return cmd
}
