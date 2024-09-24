package handler

import (
	"fmt"
	"net/http"

	"github.com/iagonc/jorge-cli/internal/schemas"

	"github.com/gin-gonic/gin"
)

// @Summary List resources by name
// @Description Retrieve a list of resources filtered by name
// @Tags resource
// @Accept json
// @Produce json
// @Param name query string true "Resource name"
// @Success 200 {array} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /resources/name [get]
func ListResourcesByNameHandler(ctx *gin.Context){
	name := ctx.Query("name")
	
	if name == ""{
		SendError(ctx, http.StatusBadRequest, "Error: missing Name in query param")
		return
	}

	var requests []schemas.Resource

	if err := db.Where("name = ?", name).Find(&requests).Error; err != nil {
		SendError(ctx, http.StatusInternalServerError, fmt.Sprintf("Error finding resources with name %s", name))
		return
	}

	SendSuccess(ctx, "list-resources-by-name", &requests)
}
