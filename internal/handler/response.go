package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
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

type Resource struct {
    Name string `json:"name"`
    Dns  string `json:"dns"`
}

type SuccessResponse struct {
	Message string `json:"message"`
	Data Resource `json:"data"`
}
