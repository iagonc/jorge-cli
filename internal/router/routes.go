package router

import (
	"github.com/gin-gonic/gin"
	"github.com/iagonc/jorge-cli/internal/handler"
)

func initializeRoutes(router *gin.Engine, h *handler.Handler, mh *handler.MonitorHandler) {
	basePath := "/api/v1"

	v1 := router.Group(basePath)
	{
		// Legacy resource endpoints
		v1.GET("/resources", h.ListResourcesHandler)
		v1.GET("/resources/name", h.ListResourcesByNameHandler)
		v1.GET("/resource", h.GetResourceByIDHandler)
		v1.POST("/resource", h.CreateResourceHandler)
		v1.PUT("/resource", h.UpdateResourceHandler)
		v1.DELETE("/resource", h.DeleteResourceHandler)

		// Monitor endpoints
		v1.GET("/monitors", mh.ListMonitors)
		v1.GET("/monitor", mh.GetMonitor)
		v1.POST("/monitor", mh.CreateMonitor)
		v1.PUT("/monitor", mh.UpdateMonitor)
		v1.DELETE("/monitor", mh.DeleteMonitor)
		v1.GET("/monitor/results", mh.GetMonitorResults)
		v1.POST("/monitor/toggle", mh.ToggleMonitor)

		// Diagnostic endpoints (one-time checks)
		v1.POST("/diagnostic", mh.RunDiagnostic)

		// Alert endpoints
		v1.GET("/alerts", mh.ListAlerts)
		v1.POST("/alert/acknowledge", mh.AcknowledgeAlert)

		// Stats and scheduler
		v1.GET("/stats", mh.GetStats)
		v1.POST("/scheduler/start", mh.StartScheduler)
		v1.POST("/scheduler/stop", mh.StopScheduler)

		// Dashboard and reports
		v1.GET("/dashboard", mh.GetDashboardData)
		v1.GET("/report", mh.GenerateReport)
	}
}
