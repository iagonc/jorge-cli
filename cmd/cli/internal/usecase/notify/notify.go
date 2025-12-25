package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// NotifyUsecase handles notifications
type NotifyUsecase struct {
	logger *zap.Logger
	client *http.Client
}

// NewNotifyUsecase creates a new NotifyUsecase
func NewNotifyUsecase(logger *zap.Logger) *NotifyUsecase {
	return &NotifyUsecase{
		logger: logger,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// LoadConfig loads notification configuration
func (u *NotifyUsecase) LoadConfig(path string) (*models.NotificationConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config models.NotificationConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SendSlack sends a notification to Slack
func (u *NotifyUsecase) SendSlack(config *models.SlackConfig, notif *models.Notification) *models.NotificationResult {
	result := &models.NotificationResult{
		Provider:  "slack",
		Timestamp: time.Now(),
	}

	// Build Slack message
	payload := map[string]interface{}{
		"text": fmt.Sprintf("*%s*\n%s", notif.Title, notif.Message),
	}

	if config.Channel != "" {
		payload["channel"] = config.Channel
	}
	if config.Username != "" {
		payload["username"] = config.Username
	}
	if config.IconEmoji != "" {
		payload["icon_emoji"] = config.IconEmoji
	}

	// Add fields as attachments
	if len(notif.Fields) > 0 || notif.Severity != "" {
		color := u.getSeverityColor(notif.Severity)
		fields := []map[string]interface{}{}

		if notif.Severity != "" {
			fields = append(fields, map[string]interface{}{
				"title": "Severity",
				"value": notif.Severity,
				"short": true,
			})
		}
		if notif.Service != "" {
			fields = append(fields, map[string]interface{}{
				"title": "Service",
				"value": notif.Service,
				"short": true,
			})
		}
		if notif.Environment != "" {
			fields = append(fields, map[string]interface{}{
				"title": "Environment",
				"value": notif.Environment,
				"short": true,
			})
		}

		for k, v := range notif.Fields {
			fields = append(fields, map[string]interface{}{
				"title": k,
				"value": v,
				"short": true,
			})
		}

		payload["attachments"] = []map[string]interface{}{
			{
				"color":  color,
				"fields": fields,
			},
		}
	}

	body, _ := json.Marshal(payload)
	resp, err := u.client.Post(config.WebhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	result.Success = resp.StatusCode >= 200 && resp.StatusCode < 300

	if !result.Success {
		result.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}

	return result
}

// SendDiscord sends a notification to Discord
func (u *NotifyUsecase) SendDiscord(config *models.DiscordConfig, notif *models.Notification) *models.NotificationResult {
	result := &models.NotificationResult{
		Provider:  "discord",
		Timestamp: time.Now(),
	}

	color := u.getDiscordColor(notif.Severity)

	payload := map[string]interface{}{
		"embeds": []map[string]interface{}{
			{
				"title":       notif.Title,
				"description": notif.Message,
				"color":       color,
				"timestamp":   notif.Timestamp.Format(time.RFC3339),
			},
		},
	}

	if config.Username != "" {
		payload["username"] = config.Username
	}
	if config.AvatarURL != "" {
		payload["avatar_url"] = config.AvatarURL
	}

	body, _ := json.Marshal(payload)
	resp, err := u.client.Post(config.WebhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	result.Success = resp.StatusCode >= 200 && resp.StatusCode < 300

	return result
}

// SendPagerDuty sends a notification to PagerDuty
func (u *NotifyUsecase) SendPagerDuty(config *models.PagerDutyConfig, notif *models.Notification) *models.NotificationResult {
	result := &models.NotificationResult{
		Provider:  "pagerduty",
		Timestamp: time.Now(),
	}

	severity := config.Severity
	if severity == "" {
		severity = notif.Severity
	}
	if severity == "" {
		severity = "warning"
	}

	payload := map[string]interface{}{
		"routing_key":  config.RoutingKey,
		"event_action": "trigger",
		"payload": map[string]interface{}{
			"summary":  notif.Title + ": " + notif.Message,
			"source":   notif.Service,
			"severity": severity,
		},
	}

	if notif.Link != "" {
		payload["links"] = []map[string]string{
			{"href": notif.Link, "text": "View Details"},
		}
	}

	body, _ := json.Marshal(payload)
	resp, err := u.client.Post("https://events.pagerduty.com/v2/enqueue", "application/json", bytes.NewReader(body))
	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	result.Success = resp.StatusCode >= 200 && resp.StatusCode < 300

	return result
}

// SendWebhook sends a notification to a generic webhook
func (u *NotifyUsecase) SendWebhook(config *models.WebhookConfig, notif *models.Notification) *models.NotificationResult {
	result := &models.NotificationResult{
		Provider:  "webhook",
		Timestamp: time.Now(),
	}

	method := config.Method
	if method == "" {
		method = "POST"
	}

	body, _ := json.Marshal(notif)
	req, err := http.NewRequest(method, config.URL, bytes.NewReader(body))
	if err != nil {
		result.Error = err.Error()
		return result
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range config.Headers {
		req.Header.Set(k, v)
	}

	resp, err := u.client.Do(req)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	result.Success = resp.StatusCode >= 200 && resp.StatusCode < 300

	return result
}

// SendAll sends notification to all configured providers
func (u *NotifyUsecase) SendAll(config *models.NotificationConfig, notif *models.Notification) *models.NotificationBatchResult {
	batch := &models.NotificationBatchResult{
		Results: []models.NotificationResult{},
	}

	if config.Slack != nil {
		result := u.SendSlack(config.Slack, notif)
		batch.Results = append(batch.Results, *result)
		if result.Success {
			batch.Succeeded++
		} else {
			batch.Failed++
		}
	}

	if config.Discord != nil {
		result := u.SendDiscord(config.Discord, notif)
		batch.Results = append(batch.Results, *result)
		if result.Success {
			batch.Succeeded++
		} else {
			batch.Failed++
		}
	}

	if config.PagerDuty != nil {
		result := u.SendPagerDuty(config.PagerDuty, notif)
		batch.Results = append(batch.Results, *result)
		if result.Success {
			batch.Succeeded++
		} else {
			batch.Failed++
		}
	}

	if config.Webhook != nil {
		result := u.SendWebhook(config.Webhook, notif)
		batch.Results = append(batch.Results, *result)
		if result.Success {
			batch.Succeeded++
		} else {
			batch.Failed++
		}
	}

	return batch
}

// QuickSlack sends a quick Slack notification using webhook URL from env or param
func (u *NotifyUsecase) QuickSlack(webhookURL, message string) *models.NotificationResult {
	if webhookURL == "" {
		webhookURL = os.Getenv("SLACK_WEBHOOK_URL")
	}

	if webhookURL == "" {
		return &models.NotificationResult{
			Provider:  "slack",
			Success:   false,
			Error:     "no webhook URL provided",
			Timestamp: time.Now(),
		}
	}

	config := &models.SlackConfig{WebhookURL: webhookURL}
	notif := &models.Notification{
		Title:     "Notification",
		Message:   message,
		Timestamp: time.Now(),
	}

	return u.SendSlack(config, notif)
}

// GenerateConfig generates a sample notification config
func (u *NotifyUsecase) GenerateConfig() string {
	return `# Notification Configuration
slack:
  webhook_url: "https://hooks.slack.com/services/xxx/yyy/zzz"
  channel: "#alerts"
  username: "Jorge Bot"
  icon_emoji: ":robot_face:"

discord:
  webhook_url: "https://discord.com/api/webhooks/xxx/yyy"
  username: "Jorge Bot"

pagerduty:
  routing_key: "your-routing-key"
  severity: "warning"

webhook:
  url: "https://your-webhook.example.com/notify"
  method: "POST"
  headers:
    Authorization: "Bearer your-token"
`
}

func (u *NotifyUsecase) getSeverityColor(severity string) string {
	switch severity {
	case "critical":
		return "#FF0000"
	case "high":
		return "#FF6600"
	case "medium", "warning":
		return "#FFCC00"
	case "low", "info":
		return "#00CC00"
	default:
		return "#808080"
	}
}

func (u *NotifyUsecase) getDiscordColor(severity string) int {
	switch severity {
	case "critical":
		return 0xFF0000
	case "high":
		return 0xFF6600
	case "medium", "warning":
		return 0xFFCC00
	case "low", "info":
		return 0x00CC00
	default:
		return 0x808080
	}
}
