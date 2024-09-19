package router

import (
	"jorge-cli/handler"

	"github.com/gin-gonic/gin"
)

func Initialize() {
	router := gin.Default()
	handler.InitializeHandler()

	initializeRoutes(router)

	router.Run(":8080")
}
