package watch

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"go.uber.org/zap"
)

// WatchUsecase handles endpoint watching
type WatchUsecase struct {
	logger *zap.Logger
}

// NewWatchUsecase creates a new WatchUsecase
func NewWatchUsecase(logger *zap.Logger) *WatchUsecase {
	return &WatchUsecase{logger: logger}
}

// EventCallback is called for each watch event
type EventCallback func(event models.WatchEvent, stats models.WatchStats)

// Watch starts watching an endpoint
func (u *WatchUsecase) Watch(ctx context.Context, config models.WatchConfig, callback EventCallback) error {
	if config.Interval == 0 {
		config.Interval = 5 * time.Second
	}
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}
	if config.ExpectedStatus == 0 {
		config.ExpectedStatus = 200
	}
	if config.Method == "" {
		config.Method = "GET"
	}

	client := &http.Client{
		Timeout: config.Timeout,
	}

	stats := models.WatchStats{
		MinResponseTime: time.Hour, // Start high
	}

	var consecutiveFailures int
	var lastDownTime time.Time
	var currentDowntimeStart time.Time

	ticker := time.NewTicker(config.Interval)
	defer ticker.Stop()

	// Initial check
	u.performCheck(ctx, client, config, &stats, &consecutiveFailures, &lastDownTime, &currentDowntimeStart, callback)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			u.performCheck(ctx, client, config, &stats, &consecutiveFailures, &lastDownTime, &currentDowntimeStart, callback)

			// Check max failures
			if config.MaxFailures > 0 && consecutiveFailures >= config.MaxFailures {
				return fmt.Errorf("max failures reached: %d consecutive failures", consecutiveFailures)
			}
		}
	}
}

func (u *WatchUsecase) performCheck(
	ctx context.Context,
	client *http.Client,
	config models.WatchConfig,
	stats *models.WatchStats,
	consecutiveFailures *int,
	lastDownTime *time.Time,
	currentDowntimeStart *time.Time,
	callback EventCallback,
) {
	event := models.WatchEvent{
		Timestamp: time.Now(),
	}

	var bodyReader io.Reader
	if config.Body != "" {
		bodyReader = strings.NewReader(config.Body)
	}

	req, err := http.NewRequestWithContext(ctx, config.Method, config.URL, bodyReader)
	if err != nil {
		event.Status = models.WatchStatusDown
		event.Error = err.Error()
		u.updateStats(stats, event, consecutiveFailures, lastDownTime, currentDowntimeStart)
		callback(event, *stats)
		return
	}

	for k, v := range config.Headers {
		req.Header.Set(k, v)
	}

	start := time.Now()
	resp, err := client.Do(req)
	event.ResponseTime = time.Since(start)

	if err != nil {
		event.Status = models.WatchStatusDown
		event.Error = err.Error()
		u.updateStats(stats, event, consecutiveFailures, lastDownTime, currentDowntimeStart)
		callback(event, *stats)
		return
	}
	defer resp.Body.Close()

	event.StatusCode = resp.StatusCode

	// Check expected status
	if resp.StatusCode != config.ExpectedStatus {
		event.Status = models.WatchStatusDown
		event.Error = fmt.Sprintf("expected %d, got %d", config.ExpectedStatus, resp.StatusCode)
	} else {
		// Check expected body if specified
		if config.ExpectedBody != "" {
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
			if strings.Contains(string(body), config.ExpectedBody) {
				event.Status = models.WatchStatusUp
				event.BodyMatch = true
			} else {
				event.Status = models.WatchStatusDown
				event.Error = "body does not match expected content"
				event.BodyMatch = false
			}
		} else {
			event.Status = models.WatchStatusUp
		}
	}

	// Check for degraded (slow response)
	if event.Status == models.WatchStatusUp && event.ResponseTime > 2*time.Second {
		event.Status = models.WatchStatusDegraded
	}

	u.updateStats(stats, event, consecutiveFailures, lastDownTime, currentDowntimeStart)
	callback(event, *stats)
}

func (u *WatchUsecase) updateStats(
	stats *models.WatchStats,
	event models.WatchEvent,
	consecutiveFailures *int,
	lastDownTime *time.Time,
	currentDowntimeStart *time.Time,
) {
	stats.TotalChecks++

	if event.Status == models.WatchStatusUp || event.Status == models.WatchStatusDegraded {
		stats.SuccessfulChecks++

		// Calculate downtime if we were down
		if !currentDowntimeStart.IsZero() {
			downtime := time.Since(*currentDowntimeStart)
			if downtime > stats.LongestDowntime {
				stats.LongestDowntime = downtime
			}
			*currentDowntimeStart = time.Time{}
		}

		if stats.CurrentStreak < 0 {
			stats.CurrentStreak = 1
		} else {
			stats.CurrentStreak++
		}
		*consecutiveFailures = 0
	} else {
		stats.FailedChecks++
		*consecutiveFailures++

		if currentDowntimeStart.IsZero() {
			*currentDowntimeStart = event.Timestamp
		}
		*lastDownTime = event.Timestamp

		if stats.CurrentStreak > 0 {
			stats.CurrentStreak = -1
		} else {
			stats.CurrentStreak--
		}
	}

	// Update response time stats
	if event.ResponseTime > 0 {
		if event.ResponseTime < stats.MinResponseTime {
			stats.MinResponseTime = event.ResponseTime
		}
		if event.ResponseTime > stats.MaxResponseTime {
			stats.MaxResponseTime = event.ResponseTime
		}
		// Simple moving average
		if stats.AvgResponseTime == 0 {
			stats.AvgResponseTime = event.ResponseTime
		} else {
			stats.AvgResponseTime = (stats.AvgResponseTime + event.ResponseTime) / 2
		}
	}

	// Calculate uptime
	if stats.TotalChecks > 0 {
		stats.Uptime = float64(stats.SuccessfulChecks) / float64(stats.TotalChecks) * 100
	}
}

// CheckOnce performs a single check
func (u *WatchUsecase) CheckOnce(ctx context.Context, config models.WatchConfig) models.WatchEvent {
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}
	if config.ExpectedStatus == 0 {
		config.ExpectedStatus = 200
	}
	if config.Method == "" {
		config.Method = "GET"
	}

	client := &http.Client{Timeout: config.Timeout}
	event := models.WatchEvent{Timestamp: time.Now()}

	req, err := http.NewRequestWithContext(ctx, config.Method, config.URL, nil)
	if err != nil {
		event.Status = models.WatchStatusDown
		event.Error = err.Error()
		return event
	}

	start := time.Now()
	resp, err := client.Do(req)
	event.ResponseTime = time.Since(start)

	if err != nil {
		event.Status = models.WatchStatusDown
		event.Error = err.Error()
		return event
	}
	defer resp.Body.Close()

	event.StatusCode = resp.StatusCode
	if resp.StatusCode == config.ExpectedStatus {
		event.Status = models.WatchStatusUp
	} else {
		event.Status = models.WatchStatusDown
		event.Error = fmt.Sprintf("expected %d, got %d", config.ExpectedStatus, resp.StatusCode)
	}

	return event
}
