package router

import (
	"github.com/gin-gonic/gin"
	"github.com/iagonc/jorge-cli/internal/handler"
)

func Initialize(h *handler.Handler) {
	router := gin.Default()

	// Inicialize as rotas com o handler injetado
	initializeRoutes(router, h)

	// Inicie o servidor
	router.Run(":8080")
}
