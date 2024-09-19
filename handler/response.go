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
		"statusCode": code,
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