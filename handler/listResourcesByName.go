package handler

import (
	"fmt"
	"jorge-cli/schemas"
	"net/http"

	"github.com/gin-gonic/gin"
)

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
