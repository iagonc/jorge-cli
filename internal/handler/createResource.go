package handler

import (
	"fmt"
	"net/http"

	"github.com/iagonc/jorge-cli/internal/schemas"

	"github.com/gin-gonic/gin"
)

// @Summary Create resource
// @Description Create a new resource with DNS
// @Tags resource
// @Accept json
// @Produce json
// @Param request body schemas.Resource true "Request body"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /resource [post]
func CreateResourceHandler(ctx *gin.Context){
	// TODO: modify this function to validate if already exist a resource with the same name
	ValidateEmptyRequest(ctx)
	if ctx.IsAborted() {
		return
	}

	var request schemas.Resource

	ValidateRequiredFields(ctx, &request)
	if ctx.IsAborted() {
		return
	}

	var existingResource schemas.Resource
	if err := db.Where("name = ?", request.Name).First(&existingResource).Error; err == nil {
		SendError(ctx, http.StatusConflict, fmt.Sprintf("Resource with name: %s already exists", request.Name))
		return
	}

	if err := db.Where("dns = ?", request.Dns).First(&existingResource).Error; err == nil {
		SendError(ctx, http.StatusConflict, fmt.Sprintf("Resource with DNS: %s already exists", request.Dns))
		return
	}

	err := db.Create(&request).Error
	if err != nil {
		SendError(ctx, http.StatusInternalServerError, "Error: could not create resource")
		return
	}

	SendSuccess(ctx, "create-resource", &request)
}
