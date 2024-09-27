package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListResourcesByNameHandler handles retrieving resources by a specified name.
func (h *Handler) ListResourcesByNameHandler(ctx *gin.Context) {
    name := ctx.Query("name")
    if name == "" {
        SendError(ctx, http.StatusBadRequest, "Error: missing name query parameter")
        return
    }

    resources, err := h.ListResourcesByNameUseCase.Execute(name)
    if err != nil {
        SendError(ctx, http.StatusInternalServerError, "Error fetching resources")
        return
    }

    SendSuccess(ctx, "list-resources-by-name", resources)
}
