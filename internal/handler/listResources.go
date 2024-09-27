package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListResourcesHandler handles the request to retrieve all resources.
func (h *Handler) ListResourcesHandler(ctx *gin.Context) {
    resources, err := h.ListResourcesUseCase.Execute()
    if err != nil {
        SendError(ctx, http.StatusInternalServerError, "Error fetching resources")
        return
    }

    SendSuccess(ctx, "list-resources", resources)
}
