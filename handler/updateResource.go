package handler

import (
	"net/http"

	"github.com/iagonc/jorge-cli/schemas"

	"github.com/gin-gonic/gin"
)

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