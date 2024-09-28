package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/iagonc/jorge-cli/internal/repository"
)

// DeleteResourceHandler handles the deletion of a resource by calling the use case.
// DeleteResourceHandler deletes a resource by ID
// @Summary Delete a resource
// @Description Delete a resource by its ID
// @Tags Resources
// @Param id query int true "Resource ID"
// @Success 200 {object} gin.H "Successfully deleted"
// @Failure 400 {object} gin.H "Bad request"
// @Failure 404 {object} gin.H "Resource not found"
// @Router /resource [delete]
func (h *Handler) GetResourceByIDHandler(ctx *gin.Context) {
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

    resource, err := h.GetResourceByID.Execute(uint(resourceID))
    if err != nil {
        if err == repository.ErrResourceNotFound {
            SendError(ctx, http.StatusNotFound, err.Error())
            return
        }

        SendError(ctx, http.StatusInternalServerError, "Error: could not get resource")
        return
    }

    SendSuccess(ctx, "get-resource-by-id", resource)
}
