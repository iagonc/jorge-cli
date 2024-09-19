package handler

import (
	"jorge-cli/schemas"
	"net/http"

	"github.com/gin-gonic/gin"
)

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
