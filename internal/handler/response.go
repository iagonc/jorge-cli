package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger precisa ser injetado corretamente
var logger *zap.Logger

// SendError envia a resposta de erro e faz log
func SendError(ctx *gin.Context, code int, message string) {
	// Definir o Content-Type já é automático com ctx.JSON
	ctx.JSON(code, gin.H{
		"message":   message,
		"errorCode": code,
	})

	// Verifica se o logger está inicializado
	logger.Sugar().Error(message)

}

// SendSuccess envia a resposta de sucesso e faz log
func SendSuccess[T any](ctx *gin.Context, op string, data T) {
	message := fmt.Sprintf("operation from handler: %s successful", op)

	ctx.JSON(http.StatusOK, gin.H{
		"message": message,
		"data":    data,
	})

	logger.Sugar().Info(message)
}

// Estruturas de resposta, se ainda forem necessárias
type ErrorResponse struct {
	Message   string `json:"message"`
	ErrorCode string `json:"errorCode"`
}

type Resource struct {
	Name string `json:"name"`
	Dns  string `json:"dns"`
}

type SuccessResponse struct {
	Message string   `json:"message"`
	Data    Resource `json:"data"`
}
