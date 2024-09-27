package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SendSuccess sends a successful response with a message and data.
func SendSuccess(ctx *gin.Context, message string, data interface{}) {
    ctx.JSON(http.StatusOK, gin.H{
        "message": message,
        "data":    data,
    })
}

// SendError sends an error response with the appropriate status code and message.
func SendError(ctx *gin.Context, statusCode int, message string) {
    ctx.JSON(statusCode, gin.H{
        "error":   message,
    })
}

