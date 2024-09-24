package handler

import (
	"fmt"
	"net/http"

	"github.com/iagonc/jorge-cli/schemas"

	"github.com/gin-gonic/gin"
)

func DeleteResourceHandler(ctx *gin.Context){
	ValidateEmptyRequest(ctx)
	if ctx.IsAborted() {
		return
	}

	id := ctx.Query("id")
	
	if id == ""{
		SendError(ctx, http.StatusBadRequest, "Error: missing ID field")
		return
	}

	var request schemas.Resource

	ValidateRequiredFields(ctx, &request)
	if ctx.IsAborted() {
		return
	}

	if err := db.First(&request, id).Error; err != nil {
		SendError(ctx, http.StatusNotFound, fmt.Sprintf("Resource with id: %s not found", id))
		return
	}

	err := db.Delete(&request).Error
	if err != nil {
		SendError(ctx, http.StatusInternalServerError, "Error: could not delete resource")
		return
	}

	SendSuccess(ctx, "delete-resource", &request)
}