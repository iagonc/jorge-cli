package models

import "time"

// NotificationConfig represents notification configuration
type NotificationConfig struct {
	Slack    *SlackConfig    `json:"slack,omitempty" yaml:"slack,omitempty"`
	Discord  *DiscordConfig  `json:"discord,omitempty" yaml:"discord,omitempty"`
	PagerDuty *PagerDutyConfig `json:"pagerduty,omitempty" yaml:"pagerduty,omitempty"`
	Webhook  *WebhookConfig  `json:"webhook,omitempty" yaml:"webhook,omitempty"`
	Email    *EmailConfig    `json:"email,omitempty" yaml:"email,omitempty"`
}

// SlackConfig represents Slack notification config
type SlackConfig struct {
	WebhookURL string `json:"webhook_url" yaml:"webhook_url"`
	Channel    string `json:"channel,omitempty" yaml:"channel,omitempty"`
	Username   string `json:"username,omitempty" yaml:"username,omitempty"`
	IconEmoji  string `json:"icon_emoji,omitempty" yaml:"icon_emoji,omitempty"`
}

// DiscordConfig represents Discord notification config
type DiscordConfig struct {
	WebhookURL string `json:"webhook_url" yaml:"webhook_url"`
	Username   string `json:"username,omitempty" yaml:"username,omitempty"`
	AvatarURL  string `json:"avatar_url,omitempty" yaml:"avatar_url,omitempty"`
}

// PagerDutyConfig represents PagerDuty notification config
type PagerDutyConfig struct {
	RoutingKey string `json:"routing_key" yaml:"routing_key"`
	Severity   string `json:"severity,omitempty" yaml:"severity,omitempty"`
}

// WebhookConfig represents generic webhook config
type WebhookConfig struct {
	URL     string            `json:"url" yaml:"url"`
	Method  string            `json:"method,omitempty" yaml:"method,omitempty"`
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
}

// EmailConfig represents email notification config
type EmailConfig struct {
	SMTPHost     string   `json:"smtp_host" yaml:"smtp_host"`
	SMTPPort     int      `json:"smtp_port" yaml:"smtp_port"`
	Username     string   `json:"username,omitempty" yaml:"username,omitempty"`
	Password     string   `json:"password,omitempty" yaml:"password,omitempty"`
	From         string   `json:"from" yaml:"from"`
	To           []string `json:"to" yaml:"to"`
	SubjectPrefix string  `json:"subject_prefix,omitempty" yaml:"subject_prefix,omitempty"`
}

// Notification represents a notification to be sent
type Notification struct {
	Title       string            `json:"title"`
	Message     string            `json:"message"`
	Severity    string            `json:"severity,omitempty"`
	Service     string            `json:"service,omitempty"`
	Environment string            `json:"environment,omitempty"`
	Link        string            `json:"link,omitempty"`
	Fields      map[string]string `json:"fields,omitempty"`
	Timestamp   time.Time         `json:"timestamp"`
}

// NotificationResult represents the result of sending a notification
type NotificationResult struct {
	Provider   string `json:"provider"`
	Success    bool   `json:"success"`
	StatusCode int    `json:"status_code,omitempty"`
	Error      string `json:"error,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}

// NotificationBatchResult represents results of sending to multiple providers
type NotificationBatchResult struct {
	Results   []NotificationResult `json:"results"`
	Succeeded int                  `json:"succeeded"`
	Failed    int                  `json:"failed"`
}
