package handler

import (
	"net/http"

	"github.com/iagonc/jorge-cli/internal/schemas"

	"github.com/gin-gonic/gin"
)

// @Summary Update resource
// @Description Update an existing resource by its ID
// @Tags resource
// @Accept json
// @Produce json
// @Param id query string true "Resource ID"
// @Param request body Resource true "Updated resource details"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /resource [put]
func UpdateResourceHandler(ctx *gin.Context){
	id := ctx.Query("id")
	if id == "" {
		SendError(ctx, http.StatusBadRequest, "ID is required")
		return
	}

	var request schemas.Resource

	ValidateRequiredFields(ctx, &request)
	if ctx.IsAborted() {
		return
	}

	// TODO: migrar pra validation
	var resource schemas.Resource
	if err := db.First(&resource, id).Error; err != nil {
		SendError(ctx, http.StatusNotFound, "Resource not found")
		return
	}

	if err := db.Model(&resource).Updates(request).Error; err != nil {
		SendError(ctx, http.StatusInternalServerError, "Error updating resource")
		return
	}

	SendSuccess(ctx, "resource-updated", &resource)
	
}