package router

import (
	"github.com/gin-gonic/gin"
	docs "github.com/iagonc/jorge-cli/docs" // Importando o Swagger gerado automaticamente
	"github.com/iagonc/jorge-cli/internal/handler"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func initializeRoutes(router *gin.Engine, h *handler.Handler) {
	basePath := "/api/v1"
	docs.SwaggerInfo.BasePath = basePath // Definindo o caminho base no Swagger

	v1 := router.Group(basePath)
	{
		v1.GET("/resources", h.ListResourcesHandler)
		v1.GET("/resources/name", h.ListResourcesByNameHandler)
		v1.POST("/resource", h.CreateResourceHandler)
		v1.PUT("/resource", h.UpdateResourceHandler)
		v1.DELETE("/resource", h.DeleteResourceHandler)
	}

	// Rota para o Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}
