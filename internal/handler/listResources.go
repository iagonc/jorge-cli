package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListResourcesHandler é responsável por chamar o use case para listar recursos
func (h *Handler) ListResourcesHandler(ctx *gin.Context) {
    resources, err := h.ListResourcesUseCase.Execute()
    if err != nil {
        SendError(ctx, http.StatusInternalServerError, "Error listing resources")
        return
    }

    SendSuccess(ctx, "list-resources", resources)
}
