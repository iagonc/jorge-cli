package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/iagonc/jorge-cli/internal/schemas"
)

// UpdateResourceHandler é responsável por chamar o use case para atualizar o recurso
func (h *Handler) UpdateResourceHandler(ctx *gin.Context) {
    id := ctx.Query("id")
    if id == "" {
        SendError(ctx, http.StatusBadRequest, "ID is required")
        return
    }

    resourceID, err := strconv.Atoi(id)
    if err != nil {
        SendError(ctx, http.StatusBadRequest, "Invalid resource ID")
        return
    }

    var request schemas.Resource
    ValidateRequiredFields(ctx, &request)
    if ctx.IsAborted() {
        return
    }

    request.ID = uint(resourceID)
    err = h.UpdateResourceUseCase.Execute(&request)
    if err != nil {
        if err.Error() == "resource not found" {
            SendError(ctx, http.StatusNotFound, "Resource not found")
        } else {
            SendError(ctx, http.StatusInternalServerError, "Error updating resource")
        }
        return
    }

    SendSuccess(ctx, "resource-updated", &request)
}
