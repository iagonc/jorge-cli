package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/iagonc/jorge-cli/internal/handler"
)

func Initialize(h *handler.Handler, mh *handler.MonitorHandler) {
	router := gin.Default()

	// CORS middleware - allow all origins for development
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
	}))

	// Initialize routes with both handlers
	initializeRoutes(router, h, mh)

	// Start the server
	router.Run(":8080")
}
