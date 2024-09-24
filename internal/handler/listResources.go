package handler

import (
	"net/http"

	"github.com/iagonc/jorge-cli/internal/schemas"

	"github.com/gin-gonic/gin"
)

// @Summary List resources
// @Description Retrieve a list of resources
// @Tags resource
// @Accept json
// @Produce json
// @Success 200 {array} Resource
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /resources [get]
func ListResourcesHandler(ctx *gin.Context){
	var request []schemas.Resource

	if err := db.Find(&request).Error; err != nil {
		SendError(ctx, http.StatusNotFound, "error listing openings")
		return
	}

	SendSuccess(ctx, "list-resource", &request)
}