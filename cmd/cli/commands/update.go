package commands

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/iagonc/jorge-cli/cmd/cli/pkg/services"

	"go.uber.org/zap"
)

func NewUpdateCommand(service *services.ResourceService) *cobra.Command {
    var id, name, dns string

    cmd := &cobra.Command{
        Use:   "update",
        Short: "Atualiza um recurso existente",
        Run: func(cmd *cobra.Command, args []string) {
            ctx := cmd.Context()
            idInt, err := strconv.Atoi(id)
            if err != nil {
                service.Logger.Error("--id deve ser um inteiro válido")
                return
            }

            updatedResource, err := service.UpdateResource(ctx, idInt, name, dns)
            if err != nil {
                service.Logger.Error("Erro ao atualizar recurso", zap.Error(err))
                return
            }

            successStyle := lipgloss.NewStyle().
                Bold(true).
                Foreground(lipgloss.Color("#FFD700")). // Cor dourada
                Padding(1, 2).
                Align(lipgloss.Center)

            result := successStyle.Render(
                fmt.Sprintf("Recurso Atualizado:\nID: %d\nName: %s\nDNS: %s",
                    updatedResource.ID, updatedResource.Name, updatedResource.Dns),
            )

            fmt.Println(result)
        },
    }

    // Adiciona flags para "id", "name" e "dns"
    cmd.Flags().StringVarP(&id, "id", "i", "", "ID do recurso (obrigatório)")
    cmd.Flags().StringVarP(&name, "name", "n", "", "Novo nome do recurso")
    cmd.Flags().StringVarP(&dns, "dns", "d", "", "Novo DNS do recurso")
    cmd.MarkFlagRequired("id")

    return cmd
}
