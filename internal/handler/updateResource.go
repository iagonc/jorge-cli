package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/iagonc/jorge-cli/internal/schemas"
)

// UpdateResourceHandler handles updating an existing resource by calling the use case.
// UpdateResourceHandler updates an existing resource
// @Summary Update a resource
// @Description Update an existing resource with the provided JSON body and ID
// @Tags Resources
// @Accept  json
// @Produce  json
// @Param id query int true "Resource ID"
// @Param resource body schemas.Resource true "Updated Resource Data"
// @Success 200 {object} schemas.Resource "Successfully updated"
// @Failure 400 {object} gin.H "Bad request"
// @Failure 404 {object} gin.H "Resource not found"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /resource [put]
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
