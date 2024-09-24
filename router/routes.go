package router

import (
	"github.com/iagonc/jorge-cli/handler"

	"github.com/gin-gonic/gin"
	docs "github.com/iagonc/jorge-cli/docs"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func initializeRoutes(router *gin.Engine) {
	basePath := "/api/v1"
	docs.SwaggerInfo.BasePath = basePath
	v1:=router.Group(basePath)
	{
		v1.GET("/resources", handler.ListResourcesHandler)
		v1.GET("/resources/name", handler.ListResourcesByNameHandler)
		v1.POST("/resource", handler.CreateResourceHandler)
		v1.PUT("/resource", handler.UpdateResourceHandler)
		v1.DELETE("/resource", handler.DeleteResourceHandler)
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}