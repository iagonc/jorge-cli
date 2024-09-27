package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListResourcesByNameHandler é responsável por chamar o use case para listar recursos por nome
func (h *Handler) ListResourcesByNameHandler(ctx *gin.Context) {
    name := ctx.Query("name")
    if name == "" {
        SendError(ctx, http.StatusBadRequest, "Error: missing name field")
        return
    }

    resources, err := h.ListResourcesByNameUseCase.Execute(name)
    if err != nil {
        SendError(ctx, http.StatusInternalServerError, "Error listing resources by name")
        return
    }

    SendSuccess(ctx, "list-resources-by-name", resources)
}
