package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/iagonc/jorge-cli/internal/schemas"
)

// ListResourcesHandler handles the request to retrieve all resources.
// ListResourcesHandler lists all resources
// @Summary List all resources
// @Description Retrieve a list of all resources
// @Tags Resources
// @Produce  json
// @Success 200 {array} schemas.Resource "List of resources"
// @Failure 500 {object} gin.H "Internal Server Error"
// @Router /resources [get]
func (h *Handler) ListResourcesHandler(ctx *gin.Context) {
    resources, err := h.ListResourcesUseCase.Execute()
    if err != nil {
        SendError(ctx, http.StatusInternalServerError, "Error fetching resources")
        return
    }

    SendSuccess(ctx, "list-resources", resources)
}
