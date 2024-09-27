package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/iagonc/jorge-cli/internal/repository"
)

// DeleteResourceHandler é responsável por chamar o use case para deletar o recurso
func (h *Handler) DeleteResourceHandler(ctx *gin.Context) {
    id := ctx.Query("id")
    if id == "" {
        SendError(ctx, http.StatusBadRequest, "Error: missing ID field")
        return
    }

    resourceID, err := strconv.Atoi(id)
    if err != nil {
        SendError(ctx, http.StatusBadRequest, "Invalid resource ID")
        return
    }

    err = h.DeleteResourceUseCase.Execute(uint(resourceID))
    if err != nil {
        // Verifica se o erro é um erro de recurso não encontrado
        if err == repository.ErrResourceNotFound {
            SendError(ctx, http.StatusNotFound, err.Error())
            return
        }

        // Para outros tipos de erro, retorna um erro interno
        SendError(ctx, http.StatusInternalServerError, "Error: could not delete resource")
        return
    }

    SendSuccess(ctx, "delete-resource", resourceID)
}
