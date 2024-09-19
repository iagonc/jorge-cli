package router

import (
	"jorge-cli/handler"

	"github.com/gin-gonic/gin"
)

func initializeRoutes(router *gin.Engine) {
	v1:=router.Group("/api/v1")
	{
		v1.GET("/resource", handler.ListResourceHandler)
		v1.GET("/resources/name", handler.ListResourcesByNameHandler)
		v1.POST("/resource", handler.CreateResourceHandler)
		v1.PUT("/resource", handler.UpdateResourceHandler)
		v1.DELETE("/resource", handler.DeleteResourceHandler)
	}
}