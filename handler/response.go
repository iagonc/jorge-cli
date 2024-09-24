package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iagonc/jorge-cli/schemas"
)

func SendError(ctx *gin.Context, code int, message string){
	ctx.Header("Content-type", "application/json")
	ctx.JSON(http.StatusBadRequest, gin.H{
		"message": message,
		"errorCode": code,
	})
	
	logger.Sugar().Error(message)
}

func SendSuccess[T any](ctx *gin.Context, op string, data *T){
	message := fmt.Sprintf("operation from handler: %s successfull", op)

	ctx.Header("Content-type", "application/json")
	ctx.JSON(http.StatusOK, gin.H{
		"message": message,
		"data": data,
	})
	
	logger.Sugar().Info(message)
}

type ErrorResponse struct {
	Message string `json:"message"`
	ErrorCode string  `json:"errorCode"`
}

type CreateResourceResponse struct {
	Message string `json:"message"`
	Data schemas.Resource `json:"data"`
}