package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/iagonc/jorge-cli/internal/schemas"
)

// ListResourcesByNameHandler handles retrieving resources by a specified name.
// ListResourcesByNameHandler lists resources by name
// @Summary List resources by name
// @Description Retrieve a list of resources that match the given name
// @Tags Resources
// @Produce  json
// @Param name query string true "Resource Name"
// @Success 200 {array} schemas.Resource "List of resources"
// @Failure 400 {object} gin.H "Bad request"
// @Failure 500 {object} gin.H "Internal Server Error"
// @Router /resources/name [get]
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
