package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"strings"
	"time"
)

// ChannelType represents the type of notification channel
type ChannelType string

const (
	Email   ChannelType = "email"
	Webhook ChannelType = "webhook"
	Slack   ChannelType = "slack"
)

// NotificationChannel represents a notification channel configuration
type NotificationChannel struct {
	ID      int         `json:"id"`
	Name    string      `json:"name"`
	Type    ChannelType `json:"type"`
	Config  string      `json:"config"` // JSON configuration
	Enabled bool        `json:"enabled"`
}

// EmailConfig represents email notification configuration
type EmailConfig struct {
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	From         string `json:"from"`
	To           string `json:"to"`
	UseTLS       bool   `json:"use_tls"`
	SkipVerify   bool   `json:"skip_verify"`
}

// WebhookConfig represents webhook notification configuration
type WebhookConfig struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
}

// SlackConfig represents Slack notification configuration
type SlackConfig struct {
	WebhookURL string `json:"webhook_url"`
	Channel    string `json:"channel"`
	Username   string `json:"username"`
}

// Alert represents an alert to be sent
type Alert struct {
	ID          int       `json:"id"`
	AlertRuleID int       `json:"alert_rule_id"`
	ServerID    int       `json:"server_id"`
	ServerName  string    `json:"server_name"`
	Metric      string    `json:"metric"`
	MetricValue float64   `json:"metric_value"`
	Threshold   float64   `json:"threshold"`
	Operator    string    `json:"operator"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
}

// Notifier handles sending notifications
type Notifier struct {
	channels []NotificationChannel
}

// NewNotifier creates a new notifier
func NewNotifier(channels []NotificationChannel) *Notifier {
	return &Notifier{
		channels: channels,
	}
}

// SendNotification sends an alert notification through all enabled channels
func (n *Notifier) SendNotification(alert Alert) error {
	for _, channel := range n.channels {
		if !channel.Enabled {
			continue
		}

		switch channel.Type {
		case Email:
			if err := n.sendEmail(channel, alert); err != nil {
				return fmt.Errorf("failed to send email notification: %v", err)
			}
		case Webhook:
			if err := n.sendWebhook(channel, alert); err != nil {
				return fmt.Errorf("failed to send webhook notification: %v", err)
			}
		case Slack:
			if err := n.sendSlack(channel, alert); err != nil {
				return fmt.Errorf("failed to send Slack notification: %v", err)
			}
		default:
			return fmt.Errorf("unsupported notification channel type: %s", channel.Type)
		}
	}

	return nil
}

// sendEmail sends an email notification
func (n *Notifier) sendEmail(channel NotificationChannel, alert Alert) error {
	var config EmailConfig
	if err := json.Unmarshal([]byte(channel.Config), &config); err != nil {
		return fmt.Errorf("failed to parse email config: %v", err)
	}

	// Create email message
	subject := fmt.Sprintf("VPS Monitor Alert: %s on %s", alert.Metric, alert.ServerName)
	body := fmt.Sprintf(`
VPS Monitor Alert Triggered

Server: %s
Metric: %s
Current Value: %.2f
Threshold: %.2f
Operator: %s
Time: %s

Message: %s
`, alert.ServerName, alert.Metric, alert.MetricValue, alert.Threshold, alert.Operator,
		alert.Timestamp.Format("2006-01-02 15:04:05"), alert.Message)

	// Create email content
	message := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", config.To, subject, body)

	// Connect to SMTP server
	auth := smtp.PlainAuth("", config.Username, config.Password, config.SMTPHost)
	addr := fmt.Sprintf("%s:%d", config.SMTPHost, config.SMTPPort)

	// Send email
	if err := smtp.SendMail(addr, auth, config.From, strings.Split(config.To, ","), []byte(message)); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}

// sendWebhook sends a webhook notification
func (n *Notifier) sendWebhook(channel NotificationChannel, alert Alert) error {
	var config WebhookConfig
	if err := json.Unmarshal([]byte(channel.Config), &config); err != nil {
		return fmt.Errorf("failed to parse webhook config: %v", err)
	}

	// Prepare payload
	payload, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert payload: %v", err)
	}

	// Create HTTP request
	method := "POST"
	if config.Method != "" {
		method = config.Method
	}

	req, err := http.NewRequest(method, config.URL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %v", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// sendSlack sends a Slack notification
func (n *Notifier) sendSlack(channel NotificationChannel, alert Alert) error {
	var config SlackConfig
	if err := json.Unmarshal([]byte(channel.Config), &config); err != nil {
		return fmt.Errorf("failed to parse Slack config: %v", err)
	}

	// Prepare Slack message payload
	slackMessage := map[string]interface{}{
		"text": fmt.Sprintf("VPS Monitor Alert: %s on %s", alert.Metric, alert.ServerName),
		"attachments": []map[string]interface{}{
			{
				"color": "danger",
				"fields": []map[string]interface{}{
					{"title": "Server", "value": alert.ServerName, "short": true},
					{"title": "Metric", "value": alert.Metric, "short": true},
					{"title": "Current Value", "value": fmt.Sprintf("%.2f", alert.MetricValue), "short": true},
					{"title": "Threshold", "value": fmt.Sprintf("%.2f %s", alert.Threshold, alert.Operator), "short": true},
					{"title": "Message", "value": alert.Message, "short": false},
					{"title": "Time", "value": alert.Timestamp.Format("2006-01-02 15:04:05"), "short": true},
				},
			},
		},
	}

	if config.Channel != "" {
		slackMessage["channel"] = config.Channel
	}

	if config.Username != "" {
		slackMessage["username"] = config.Username
	}

	// Convert to JSON
	payload, err := json.Marshal(slackMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack message: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", config.WebhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create Slack request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Slack notification: %v", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Slack webhook returned status %d", resp.StatusCode)
	}

	return nil
}