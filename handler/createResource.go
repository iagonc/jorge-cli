package handler

import (
	"net/http"

	"github.com/iagonc/jorge-cli/schemas"

	"github.com/gin-gonic/gin"
)

// @BasePath /api/v1

// @Summary Create resource
// @Description Create a new resource with DNS
// @Tags resource
// @Accept json
// @Produce json
// @Param request body schemas.Resource true "Request body"
// @Success 200 {object} CreateResourceResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /resource [post]

func CreateResourceHandler(ctx *gin.Context){
	ValidateEmptyRequest(ctx)
	if ctx.IsAborted() {
		return
	}

	var request schemas.Resource

	ValidateRequiredFields(ctx, &request)
	if ctx.IsAborted() {
		return
	}

	err := db.Create(&request).Error
	if err != nil {
		SendError(ctx, http.StatusInternalServerError, "Error: could not create resource")
		return
	}

	SendSuccess(ctx, "create-resource", &request)
}
