package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func ValidateEmptyRequest(ctx *gin.Context) {
	if ctx.Request.Body == nil || ctx.Request.ContentLength == 0{
		SendError(ctx, http.StatusBadRequest, "Error: request body is empty")

		ctx.Abort()
	}
}

func ValidateRequiredFields[T any](ctx *gin.Context, request *T){
	validate := validator.New()

	ctx.BindJSON(request)
	
    if err := validate.Struct(request); err != nil {
        SendError(ctx, http.StatusBadRequest, fmt.Sprintf("Error: %s", err.Error()))
        ctx.Abort()
    }
}
