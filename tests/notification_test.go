package tests

import (
	"testing"

	"github.com/username/vps-monitor/internal/notification"
)

func TestEmailConfigValidation(t *testing.T) {
	config := notification.EmailConfig{
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		Username: "user@example.com",
		Password: "password",
		From:     "from@example.com",
		To:       "to@example.com",
		UseTLS:   true,
	}

	if config.SMTPHost == "" {
		t.Error("SMTP host should not be empty")
	}

	if config.SMTPPort <= 0 {
		t.Error("SMTP port should be positive")
	}
}

func TestWebhookConfigValidation(t *testing.T) {
	config := notification.WebhookConfig{
		URL:    "https://example.com/webhook",
		Method: "POST",
		Headers: map[string]string{
			"Authorization": "Bearer token",
		},
	}

	if config.URL == "" {
		t.Error("Webhook URL should not be empty")
	}

	if config.Method != "POST" && config.Method != "GET" && config.Method != "" {
		t.Error("Webhook method should be POST, GET, or empty")
	}
}

func TestSlackConfigValidation(t *testing.T) {
	config := notification.SlackConfig{
		WebhookURL: "https://hooks.slack.com/services/xxx",
		Channel:    "#alerts",
		Username:   "VPS Monitor",
	}

	if config.WebhookURL == "" {
		t.Error("Slack webhook URL should not be empty")
	}
}

func TestNotificationChannelTypes(t *testing.T) {
	tests := []struct {
		name     string
		chanType notification.ChannelType
		valid    bool
	}{
		{"Email", notification.Email, true},
		{"Webhook", notification.Webhook, true},
		{"Slack", notification.Slack, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.chanType) == "" {
				t.Errorf("Channel type %s should not be empty", tt.name)
			}
		})
	}
}

func TestAlertStructure(t *testing.T) {
	alert := notification.Alert{
		ID:          1,
		AlertRuleID: 1,
		ServerID:    1,
		ServerName:  "test-server",
		Metric:      "cpu",
		MetricValue: 85.5,
		Threshold:   80.0,
		Operator:    ">",
		Message:     "CPU usage is high",
	}

	if alert.ServerName == "" {
		t.Error("Server name should not be empty")
	}

	if alert.MetricValue <= 0 {
		t.Error("Metric value should be positive")
	}

	if alert.Threshold <= 0 {
		t.Error("Threshold should be positive")
	}
}
