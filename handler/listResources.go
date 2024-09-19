package handler

import (
	"jorge-cli/schemas"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ListResourcesHandler(ctx *gin.Context){
	var request []schemas.Resource

	if err := db.Find(&request).Error; err != nil {
		SendError(ctx, http.StatusNotFound, "error listing openings")
		return
	}

	SendSuccess(ctx, "list-resource", &request)
}