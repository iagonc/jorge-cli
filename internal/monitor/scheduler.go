package monitor

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/iagonc/jorge-cli/internal/schemas"
	"gorm.io/gorm"
)

// Scheduler manages continuous monitoring of endpoints
type Scheduler struct {
	db       *gorm.DB
	tools    *NetworkTools
	stopChan chan struct{}
	wg       sync.WaitGroup
	running  bool
	mu       sync.RWMutex
}

// NewScheduler creates a new monitoring scheduler
func NewScheduler(db *gorm.DB) *Scheduler {
	return &Scheduler{
		db:       db,
		tools:    NewNetworkTools(),
		stopChan: make(chan struct{}),
	}
}

// Start begins the monitoring scheduler
func (s *Scheduler) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.stopChan = make(chan struct{})
	s.mu.Unlock()

	log.Println("[Scheduler] Starting monitoring scheduler...")

	s.wg.Add(1)
	go s.runLoop()
}

// Stop stops the monitoring scheduler
func (s *Scheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	close(s.stopChan)
	s.mu.Unlock()

	s.wg.Wait()
	log.Println("[Scheduler] Monitoring scheduler stopped")
}

// IsRunning returns whether the scheduler is running
func (s *Scheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

func (s *Scheduler) runLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(5 * time.Second) // Check every 5 seconds for due monitors
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.checkDueMonitors()
		}
	}
}

func (s *Scheduler) checkDueMonitors() {
	var monitors []schemas.Monitor

	// Find enabled monitors that are due for a check
	now := time.Now()
	err := s.db.Where("enabled = ?", true).Find(&monitors).Error
	if err != nil {
		log.Printf("[Scheduler] Error fetching monitors: %v", err)
		return
	}

	for _, monitor := range monitors {
		// Check if monitor is due
		if monitor.LastCheckAt == nil || now.Sub(*monitor.LastCheckAt) >= time.Duration(monitor.Interval)*time.Second {
			go s.executeCheck(&monitor)
		}
	}
}

func (s *Scheduler) executeCheck(monitor *schemas.Monitor) {
	log.Printf("[Scheduler] Checking monitor: %s (%s)", monitor.Name, monitor.Target)

	var result *schemas.DiagnosticResult

	switch monitor.Type {
	case schemas.MonitorTypeHTTP:
		result = s.tools.HTTP(monitor.Target, monitor.Method, monitor.ExpectedStatus, monitor.Timeout)
	case schemas.MonitorTypeTCP:
		result = s.tools.TCP(monitor.Target, monitor.Port, monitor.Timeout)
	case schemas.MonitorTypePing:
		result = s.tools.Ping(monitor.Target, monitor.Timeout)
	case schemas.MonitorTypeDNS:
		result = s.tools.DNS(monitor.Target, monitor.Timeout)
	case schemas.MonitorTypeSSL:
		result = s.tools.SSL(monitor.Target, monitor.Port, monitor.Timeout)
	case schemas.MonitorTypeTrace:
		result = s.tools.Traceroute(monitor.Target, monitor.Timeout)
	default:
		log.Printf("[Scheduler] Unknown monitor type: %s", monitor.Type)
		return
	}

	// Determine status
	var status schemas.MonitorStatus
	if result.Success {
		status = schemas.MonitorStatusUp
	} else {
		status = schemas.MonitorStatusDown
	}

	// Check for degraded (high latency)
	if result.Success && result.Latency > 1000 { // >1s latency
		status = schemas.MonitorStatusDegraded
	}

	// Save check result
	now := time.Now()
	checkResult := schemas.CheckResult{
		MonitorID:    monitor.ID,
		Status:       status,
		Latency:      result.Latency,
		Output:       result.Output,
		ErrorMessage: result.Error,
		CheckedAt:    now,
	}

	// Add HTTP-specific data
	if statusCode, ok := result.Details["status_code"].(int); ok {
		checkResult.StatusCode = statusCode
	}

	// Add DNS-specific data
	if ips, ok := result.Details["resolved_ips"].([]string); ok {
		checkResult.ResolvedIPs = joinStrings(ips, ",")
	}

	// Add SSL-specific data
	if notAfter, ok := result.Details["not_after"].(string); ok {
		if t, err := time.Parse(time.RFC3339, notAfter); err == nil {
			checkResult.CertExpiry = &t
		}
	}
	if issuer, ok := result.Details["issuer"].(string); ok {
		checkResult.CertIssuer = issuer
	}
	if daysLeft, ok := result.Details["days_left"].(int); ok {
		checkResult.CertDaysLeft = daysLeft
	}

	// Save result
	if err := s.db.Create(&checkResult).Error; err != nil {
		log.Printf("[Scheduler] Error saving check result: %v", err)
	}

	// Update monitor status
	previousStatus := monitor.Status
	monitor.Status = status
	monitor.LastCheckAt = &now
	monitor.LastLatency = result.Latency

	// Calculate uptime (last 100 checks)
	var successCount int64
	var totalCount int64
	s.db.Model(&schemas.CheckResult{}).
		Where("monitor_id = ? AND status = ?", monitor.ID, schemas.MonitorStatusUp).
		Order("created_at DESC").
		Limit(100).
		Count(&successCount)
	s.db.Model(&schemas.CheckResult{}).
		Where("monitor_id = ?", monitor.ID).
		Order("created_at DESC").
		Limit(100).
		Count(&totalCount)

	if totalCount > 0 {
		monitor.Uptime = float64(successCount) / float64(totalCount) * 100
	}

	if err := s.db.Save(monitor).Error; err != nil {
		log.Printf("[Scheduler] Error updating monitor: %v", err)
	}

	// Generate alert if status changed
	if previousStatus != schemas.MonitorStatusUnknown && previousStatus != status {
		s.generateAlert(monitor, previousStatus, status, result)
	}

	// Generate alert for SSL expiry warning
	if monitor.Type == schemas.MonitorTypeSSL {
		if daysLeft, ok := result.Details["days_left"].(int); ok && daysLeft <= 30 && daysLeft > 0 {
			s.generateSSLExpiryAlert(monitor, daysLeft)
		}
	}

	log.Printf("[Scheduler] Monitor %s: %s (latency: %dms)", monitor.Name, status, result.Latency)
}

func (s *Scheduler) generateAlert(monitor *schemas.Monitor, previousStatus, newStatus schemas.MonitorStatus, result *schemas.DiagnosticResult) {
	var severity schemas.AlertSeverity
	var title, message string

	if newStatus == schemas.MonitorStatusDown {
		severity = schemas.AlertSeverityCritical
		title = fmt.Sprintf("%s is DOWN", monitor.Name)
		message = fmt.Sprintf("Monitor %s (%s) changed from %s to %s. Error: %s",
			monitor.Name, monitor.Target, previousStatus, newStatus, result.Error)
	} else if newStatus == schemas.MonitorStatusDegraded {
		severity = schemas.AlertSeverityWarning
		title = fmt.Sprintf("%s is degraded", monitor.Name)
		message = fmt.Sprintf("Monitor %s (%s) is experiencing high latency: %dms",
			monitor.Name, monitor.Target, result.Latency)
	} else if newStatus == schemas.MonitorStatusUp && previousStatus == schemas.MonitorStatusDown {
		severity = schemas.AlertSeverityInfo
		title = fmt.Sprintf("%s is UP", monitor.Name)
		message = fmt.Sprintf("Monitor %s (%s) has recovered", monitor.Name, monitor.Target)
	} else {
		return
	}

	alert := schemas.Alert{
		MonitorID: monitor.ID,
		Severity:  severity,
		Title:     title,
		Message:   message,
	}

	if err := s.db.Create(&alert).Error; err != nil {
		log.Printf("[Scheduler] Error creating alert: %v", err)
	} else {
		log.Printf("[Scheduler] Alert created: %s", title)
	}
}

func (s *Scheduler) generateSSLExpiryAlert(monitor *schemas.Monitor, daysLeft int) {
	// Check if we already have a recent alert for this
	var existingAlert schemas.Alert
	oneDay := time.Now().Add(-24 * time.Hour)
	err := s.db.Where("monitor_id = ? AND title LIKE ? AND created_at > ?",
		monitor.ID, "%SSL certificate%", oneDay).First(&existingAlert).Error

	if err == nil {
		// Alert already exists
		return
	}

	severity := schemas.AlertSeverityWarning
	if daysLeft <= 7 {
		severity = schemas.AlertSeverityCritical
	}

	alert := schemas.Alert{
		MonitorID: monitor.ID,
		Severity:  severity,
		Title:     fmt.Sprintf("SSL certificate expires in %d days", daysLeft),
		Message:   fmt.Sprintf("The SSL certificate for %s will expire in %d days", monitor.Target, daysLeft),
	}

	if err := s.db.Create(&alert).Error; err != nil {
		log.Printf("[Scheduler] Error creating SSL alert: %v", err)
	}
}

func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

