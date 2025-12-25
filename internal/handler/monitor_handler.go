package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iagonc/jorge-cli/internal/monitor"
	"github.com/iagonc/jorge-cli/internal/schemas"
	"gorm.io/gorm"
)

// MonitorHandler handles monitor-related API requests
type MonitorHandler struct {
	db        *gorm.DB
	scheduler *monitor.Scheduler
	tools     *monitor.NetworkTools
}

// NewMonitorHandler creates a new MonitorHandler
func NewMonitorHandler(db *gorm.DB, scheduler *monitor.Scheduler) *MonitorHandler {
	return &MonitorHandler{
		db:        db,
		scheduler: scheduler,
		tools:     monitor.NewNetworkTools(),
	}
}

// ListMonitors returns all monitors
func (h *MonitorHandler) ListMonitors(c *gin.Context) {
	var monitors []schemas.Monitor
	if err := h.db.Order("created_at DESC").Find(&monitors).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": monitors, "message": "list-monitors"})
}

// GetMonitor returns a single monitor with recent results
func (h *MonitorHandler) GetMonitor(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var m schemas.Monitor
	if err := h.db.First(&m, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "monitor not found"})
		return
	}

	// Get last 20 results
	var results []schemas.CheckResult
	h.db.Where("monitor_id = ?", id).Order("created_at DESC").Limit(20).Find(&results)
	m.Results = results

	c.JSON(http.StatusOK, gin.H{"data": m, "message": "get-monitor"})
}

// CreateMonitor creates a new monitor
func (h *MonitorHandler) CreateMonitor(c *gin.Context) {
	var input struct {
		Name           string `json:"name" binding:"required"`
		Target         string `json:"target" binding:"required"`
		Type           string `json:"type" binding:"required"`
		Port           int    `json:"port"`
		Interval       int    `json:"interval"`
		Timeout        int    `json:"timeout"`
		ExpectedStatus int    `json:"expected_status"`
		Method         string `json:"method"`
		Enabled        *bool  `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults
	if input.Interval == 0 {
		input.Interval = 30
	}
	if input.Timeout == 0 {
		input.Timeout = 10
	}
	if input.Method == "" {
		input.Method = "GET"
	}
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}

	m := schemas.Monitor{
		Name:           input.Name,
		Target:         input.Target,
		Type:           schemas.MonitorType(input.Type),
		Port:           input.Port,
		Interval:       input.Interval,
		Timeout:        input.Timeout,
		ExpectedStatus: input.ExpectedStatus,
		Method:         input.Method,
		Enabled:        enabled,
		Status:         schemas.MonitorStatusUnknown,
	}

	if err := h.db.Create(&m).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": m, "message": "create-monitor"})
}

// UpdateMonitor updates a monitor
func (h *MonitorHandler) UpdateMonitor(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var m schemas.Monitor
	if err := h.db.First(&m, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "monitor not found"})
		return
	}

	var input map[string]interface{}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.Model(&m).Updates(input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.db.First(&m, id)
	c.JSON(http.StatusOK, gin.H{"data": m, "message": "update-monitor"})
}

// DeleteMonitor deletes a monitor
func (h *MonitorHandler) DeleteMonitor(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// Delete associated results and alerts first
	h.db.Where("monitor_id = ?", id).Delete(&schemas.CheckResult{})
	h.db.Where("monitor_id = ?", id).Delete(&schemas.Alert{})

	if err := h.db.Delete(&schemas.Monitor{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "delete-monitor"})
}

// GetMonitorResults returns recent results for a monitor
func (h *MonitorHandler) GetMonitorResults(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	var results []schemas.CheckResult
	h.db.Where("monitor_id = ?", id).Order("created_at DESC").Limit(limit).Find(&results)

	c.JSON(http.StatusOK, gin.H{"data": results, "message": "monitor-results"})
}

// ToggleMonitor enables/disables a monitor
func (h *MonitorHandler) ToggleMonitor(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var m schemas.Monitor
	if err := h.db.First(&m, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "monitor not found"})
		return
	}

	m.Enabled = !m.Enabled
	h.db.Save(&m)

	c.JSON(http.StatusOK, gin.H{"data": m, "message": "toggle-monitor"})
}

// RunDiagnostic runs a one-time diagnostic check
func (h *MonitorHandler) RunDiagnostic(c *gin.Context) {
	var input schemas.DiagnosticRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Timeout == 0 {
		input.Timeout = 10
	}

	var result *schemas.DiagnosticResult

	switch input.Type {
	case "ping":
		result = h.tools.Ping(input.Target, input.Timeout)
	case "dns":
		result = h.tools.DNS(input.Target, input.Timeout)
	case "tcp":
		if input.Port == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "port is required for TCP check"})
			return
		}
		result = h.tools.TCP(input.Target, input.Port, input.Timeout)
	case "http":
		result = h.tools.HTTP(input.Target, "GET", 0, input.Timeout)
	case "ssl":
		result = h.tools.SSL(input.Target, input.Port, input.Timeout)
	case "trace":
		result = h.tools.Traceroute(input.Target, input.Timeout)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid diagnostic type"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result, "message": "diagnostic"})
}

// ListAlerts returns all alerts
func (h *MonitorHandler) ListAlerts(c *gin.Context) {
	var alerts []schemas.Alert

	query := h.db.Order("created_at DESC").Limit(100)

	// Filter by acknowledged status
	if acked := c.Query("acknowledged"); acked != "" {
		query = query.Where("acknowledged = ?", acked == "true")
	}

	// Filter by severity
	if severity := c.Query("severity"); severity != "" {
		query = query.Where("severity = ?", severity)
	}

	if err := query.Preload("Monitor").Find(&alerts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": alerts, "message": "list-alerts"})
}

// AcknowledgeAlert acknowledges an alert
func (h *MonitorHandler) AcknowledgeAlert(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var alert schemas.Alert
	if err := h.db.First(&alert, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "alert not found"})
		return
	}

	now := time.Now()
	alert.Acknowledged = true
	alert.AckedAt = &now
	h.db.Save(&alert)

	c.JSON(http.StatusOK, gin.H{"data": alert, "message": "acknowledge-alert"})
}

// GetStats returns dashboard statistics
func (h *MonitorHandler) GetStats(c *gin.Context) {
	var totalMonitors int64
	var upMonitors int64
	var downMonitors int64
	var degradedMonitors int64
	var totalAlerts int64
	var unresolvedAlerts int64

	h.db.Model(&schemas.Monitor{}).Where("enabled = ?", true).Count(&totalMonitors)
	h.db.Model(&schemas.Monitor{}).Where("enabled = ? AND status = ?", true, schemas.MonitorStatusUp).Count(&upMonitors)
	h.db.Model(&schemas.Monitor{}).Where("enabled = ? AND status = ?", true, schemas.MonitorStatusDown).Count(&downMonitors)
	h.db.Model(&schemas.Monitor{}).Where("enabled = ? AND status = ?", true, schemas.MonitorStatusDegraded).Count(&degradedMonitors)
	h.db.Model(&schemas.Alert{}).Count(&totalAlerts)
	h.db.Model(&schemas.Alert{}).Where("acknowledged = ?", false).Count(&unresolvedAlerts)

	// Calculate average latency from last hour
	var avgLatency float64
	h.db.Model(&schemas.CheckResult{}).
		Where("created_at > ? AND status = ?", time.Now().Add(-1*time.Hour), schemas.MonitorStatusUp).
		Select("COALESCE(AVG(latency), 0)").
		Scan(&avgLatency)

	// Calculate overall uptime
	var uptime float64 = 100
	if totalMonitors > 0 {
		uptime = float64(upMonitors) / float64(totalMonitors) * 100
	}

	stats := gin.H{
		"total_monitors":     totalMonitors,
		"up_monitors":        upMonitors,
		"down_monitors":      downMonitors,
		"degraded_monitors":  degradedMonitors,
		"total_alerts":       totalAlerts,
		"unresolved_alerts":  unresolvedAlerts,
		"avg_latency_ms":     avgLatency,
		"uptime_percent":     uptime,
		"scheduler_running":  h.scheduler.IsRunning(),
	}

	c.JSON(http.StatusOK, gin.H{"data": stats, "message": "stats"})
}

// StartScheduler starts the monitoring scheduler
func (h *MonitorHandler) StartScheduler(c *gin.Context) {
	h.scheduler.Start()
	c.JSON(http.StatusOK, gin.H{"message": "scheduler-started"})
}

// StopScheduler stops the monitoring scheduler
func (h *MonitorHandler) StopScheduler(c *gin.Context) {
	h.scheduler.Stop()
	c.JSON(http.StatusOK, gin.H{"message": "scheduler-stopped"})
}

// GetDashboardData returns comprehensive dashboard data with charts
func (h *MonitorHandler) GetDashboardData(c *gin.Context) {
	// Get latency history for last 24 hours (hourly buckets)
	type LatencyBucket struct {
		Hour       string  `json:"hour"`
		AvgLatency float64 `json:"avg_latency"`
		P95Latency float64 `json:"p95_latency"`
		CheckCount int64   `json:"check_count"`
	}

	var latencyHistory []LatencyBucket
	h.db.Model(&schemas.CheckResult{}).
		Select("strftime('%Y-%m-%d %H:00', created_at) as hour, AVG(latency) as avg_latency, MAX(latency) as p95_latency, COUNT(*) as check_count").
		Where("created_at > ? AND status = ?", time.Now().Add(-24*time.Hour), schemas.MonitorStatusUp).
		Group("hour").
		Order("hour ASC").
		Scan(&latencyHistory)

	// Get uptime history per monitor for last 24 hours
	type UptimeSlot struct {
		MonitorID   uint   `json:"monitor_id"`
		MonitorName string `json:"monitor_name"`
		Hour        string `json:"hour"`
		UpCount     int64  `json:"up_count"`
		TotalCount  int64  `json:"total_count"`
	}

	var uptimeSlots []UptimeSlot
	h.db.Table("check_results").
		Select("check_results.monitor_id, monitors.name as monitor_name, strftime('%Y-%m-%d %H:00', check_results.created_at) as hour, SUM(CASE WHEN check_results.status = 'up' THEN 1 ELSE 0 END) as up_count, COUNT(*) as total_count").
		Joins("JOIN monitors ON monitors.id = check_results.monitor_id").
		Where("check_results.created_at > ?", time.Now().Add(-24*time.Hour)).
		Group("check_results.monitor_id, hour").
		Order("hour ASC").
		Scan(&uptimeSlots)

	// Get monitor status summary
	var monitors []schemas.Monitor
	h.db.Where("enabled = ?", true).Order("status ASC, name ASC").Find(&monitors)

	// Get recent incidents (status changes to down)
	type Incident struct {
		ID          uint      `json:"id"`
		MonitorID   uint      `json:"monitor_id"`
		MonitorName string    `json:"monitor_name"`
		StartedAt   time.Time `json:"started_at"`
		Status      string    `json:"status"`
		Message     string    `json:"message"`
	}

	var incidents []Incident
	h.db.Table("alerts").
		Select("alerts.id, alerts.monitor_id, monitors.name as monitor_name, alerts.created_at as started_at, alerts.severity as status, alerts.message").
		Joins("JOIN monitors ON monitors.id = alerts.monitor_id").
		Where("alerts.severity IN ('critical', 'warning')").
		Order("alerts.created_at DESC").
		Limit(10).
		Scan(&incidents)

	// Calculate SLO metrics (target 99.9% uptime)
	var totalChecks int64
	var successChecks int64
	h.db.Model(&schemas.CheckResult{}).Where("created_at > ?", time.Now().Add(-24*time.Hour)).Count(&totalChecks)
	h.db.Model(&schemas.CheckResult{}).Where("created_at > ? AND status = ?", time.Now().Add(-24*time.Hour), schemas.MonitorStatusUp).Count(&successChecks)

	var sloTarget float64 = 99.9
	var currentSLO float64 = 100
	if totalChecks > 0 {
		currentSLO = float64(successChecks) / float64(totalChecks) * 100
	}

	errorBudget := currentSLO - sloTarget
	if errorBudget < 0 {
		errorBudget = 0
	}

	// Response time percentiles
	type Percentiles struct {
		P50 float64 `json:"p50"`
		P95 float64 `json:"p95"`
		P99 float64 `json:"p99"`
	}

	var p50, p95, p99 float64
	h.db.Model(&schemas.CheckResult{}).
		Where("created_at > ? AND status = ?", time.Now().Add(-1*time.Hour), schemas.MonitorStatusUp).
		Select("COALESCE(AVG(latency), 0)").Scan(&p50)
	h.db.Model(&schemas.CheckResult{}).
		Where("created_at > ? AND status = ?", time.Now().Add(-1*time.Hour), schemas.MonitorStatusUp).
		Order("latency DESC").Limit(1).Offset(int(float64(totalChecks) * 0.05)).
		Select("COALESCE(latency, 0)").Scan(&p95)
	h.db.Model(&schemas.CheckResult{}).
		Where("created_at > ? AND status = ?", time.Now().Add(-1*time.Hour), schemas.MonitorStatusUp).
		Order("latency DESC").Limit(1).
		Select("COALESCE(latency, 0)").Scan(&p99)

	data := gin.H{
		"latency_history": latencyHistory,
		"uptime_slots":    uptimeSlots,
		"monitors":        monitors,
		"incidents":       incidents,
		"slo": gin.H{
			"target":       sloTarget,
			"current":      currentSLO,
			"error_budget": errorBudget,
			"compliant":    currentSLO >= sloTarget,
		},
		"percentiles": gin.H{
			"p50": p50,
			"p95": p95,
			"p99": p99,
		},
	}

	c.JSON(http.StatusOK, gin.H{"data": data, "message": "dashboard"})
}

// GenerateReport generates a comprehensive monitoring report
func (h *MonitorHandler) GenerateReport(c *gin.Context) {
	reportType := c.Query("type") // uptime, incident, performance
	period := c.Query("period")   // 24h, 7d, 30d

	if period == "" {
		period = "24h"
	}

	var duration time.Duration
	switch period {
	case "7d":
		duration = 7 * 24 * time.Hour
	case "30d":
		duration = 30 * 24 * time.Hour
	default:
		duration = 24 * time.Hour
	}

	startTime := time.Now().Add(-duration)

	switch reportType {
	case "incident":
		var alerts []schemas.Alert
		h.db.Preload("Monitor").Where("created_at > ?", startTime).Order("created_at DESC").Find(&alerts)

		report := gin.H{
			"type":        "incident",
			"period":      period,
			"generated":   time.Now(),
			"total":       len(alerts),
			"critical":    0,
			"warning":     0,
			"resolved":    0,
			"incidents":   alerts,
		}

		for _, a := range alerts {
			if a.Severity == schemas.AlertSeverityCritical {
				report["critical"] = report["critical"].(int) + 1
			} else if a.Severity == schemas.AlertSeverityWarning {
				report["warning"] = report["warning"].(int) + 1
			}
			if a.Acknowledged {
				report["resolved"] = report["resolved"].(int) + 1
			}
		}

		c.JSON(http.StatusOK, gin.H{"data": report, "message": "report"})

	case "performance":
		type MonitorPerf struct {
			ID          uint    `json:"id"`
			Name        string  `json:"name"`
			Target      string  `json:"target"`
			Type        string  `json:"type"`
			AvgLatency  float64 `json:"avg_latency"`
			MinLatency  float64 `json:"min_latency"`
			MaxLatency  float64 `json:"max_latency"`
			CheckCount  int64   `json:"check_count"`
			SuccessRate float64 `json:"success_rate"`
		}

		var perfData []MonitorPerf
		h.db.Table("monitors").
			Select(`
				monitors.id, monitors.name, monitors.target, monitors.type,
				COALESCE(AVG(check_results.latency), 0) as avg_latency,
				COALESCE(MIN(check_results.latency), 0) as min_latency,
				COALESCE(MAX(check_results.latency), 0) as max_latency,
				COUNT(check_results.id) as check_count,
				COALESCE(SUM(CASE WHEN check_results.status = 'up' THEN 1.0 ELSE 0.0 END) / NULLIF(COUNT(check_results.id), 0) * 100, 0) as success_rate
			`).
			Joins("LEFT JOIN check_results ON check_results.monitor_id = monitors.id AND check_results.created_at > ?", startTime).
			Where("monitors.enabled = ?", true).
			Group("monitors.id").
			Order("avg_latency DESC").
			Scan(&perfData)

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"type":      "performance",
				"period":    period,
				"generated": time.Now(),
				"monitors":  perfData,
			},
			"message": "report",
		})

	default: // uptime report
		type MonitorUptime struct {
			ID           uint    `json:"id"`
			Name         string  `json:"name"`
			Target       string  `json:"target"`
			Type         string  `json:"type"`
			TotalChecks  int64   `json:"total_checks"`
			SuccessCount int64   `json:"success_count"`
			FailCount    int64   `json:"fail_count"`
			UptimePercent float64 `json:"uptime_percent"`
		}

		var uptimeData []MonitorUptime
		h.db.Table("monitors").
			Select(`
				monitors.id, monitors.name, monitors.target, monitors.type,
				COUNT(check_results.id) as total_checks,
				SUM(CASE WHEN check_results.status = 'up' THEN 1 ELSE 0 END) as success_count,
				SUM(CASE WHEN check_results.status != 'up' THEN 1 ELSE 0 END) as fail_count,
				COALESCE(SUM(CASE WHEN check_results.status = 'up' THEN 1.0 ELSE 0.0 END) / NULLIF(COUNT(check_results.id), 0) * 100, 100) as uptime_percent
			`).
			Joins("LEFT JOIN check_results ON check_results.monitor_id = monitors.id AND check_results.created_at > ?", startTime).
			Where("monitors.enabled = ?", true).
			Group("monitors.id").
			Order("uptime_percent ASC").
			Scan(&uptimeData)

		// Calculate overall stats
		var totalChecks, successChecks int64
		for _, m := range uptimeData {
			totalChecks += m.TotalChecks
			successChecks += m.SuccessCount
		}

		var overallUptime float64 = 100
		if totalChecks > 0 {
			overallUptime = float64(successChecks) / float64(totalChecks) * 100
		}

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"type":           "uptime",
				"period":         period,
				"generated":      time.Now(),
				"overall_uptime": overallUptime,
				"total_checks":   totalChecks,
				"monitors":       uptimeData,
			},
			"message": "report",
		})
	}
}
