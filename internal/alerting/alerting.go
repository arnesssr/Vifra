package alerting

import (
	"fmt"
	"log"
	"time"

	"github.com/username/vps-monitor/internal/database"
	"github.com/username/vps-monitor/internal/models"
	"github.com/username/vps-monitor/internal/notification"
)

// AlertEvaluator evaluates metrics against alert rules and triggers notifications
type AlertEvaluator struct {
	db       *database.DB
	notifier *notification.Notifier
}

// NewAlertEvaluator creates a new alert evaluator
func NewAlertEvaluator(db *database.DB) *AlertEvaluator {
	return &AlertEvaluator{
		db: db,
	}
}

// EvaluateMetrics checks if the given metrics violate any alert rules
func (ae *AlertEvaluator) EvaluateMetrics(metrics *models.ServerMetrics) error {
	// Get server info
	var server models.Server
	if err := ae.db.First(&server, metrics.ServerID).Error; err != nil {
		return fmt.Errorf("failed to get server: %v", err)
	}

	// Get all enabled alert rules for this server and global rules
	var alertRules []models.AlertRule
	if err := ae.db.Where("enabled = ? AND (server_id = ? OR server_id IS NULL)", true, metrics.ServerID).Find(&alertRules).Error; err != nil {
		return fmt.Errorf("failed to get alert rules: %v", err)
	}

	// Evaluate each alert rule
	for _, rule := range alertRules {
		triggered, value := ae.evaluateRule(&rule, metrics)
		if triggered {
			// Create alert
			alert := &models.Alert{
				AlertRuleID: rule.ID,
				ServerID:    metrics.ServerID,
				MetricValue: value,
				Message:     fmt.Sprintf("Alert '%s' triggered: %s %s %.2f (current: %.2f)", rule.Name, rule.Metric, rule.Operator, rule.Threshold, value),
				Status:      "active",
			}

			if err := ae.db.Create(alert).Error; err != nil {
				log.Printf("Failed to create alert: %v", err)
				continue
			}

			log.Printf("Alert triggered: %s on server %s", rule.Name, server.Name)

			// Send notifications
			if err := ae.sendNotifications(alert, &server, &rule); err != nil {
				log.Printf("Failed to send notifications for alert %d: %v", alert.ID, err)
			}
		}
	}

	return nil
}

// evaluateRule checks if a single alert rule is triggered by the metrics
func (ae *AlertEvaluator) evaluateRule(rule *models.AlertRule, metrics *models.ServerMetrics) (bool, float64) {
	var value float64

	// Extract the metric value based on the rule's metric type
	switch rule.Metric {
	case "cpu":
		value = metrics.CPUUsage
	case "memory":
		if metrics.MemoryTotal > 0 {
			value = (float64(metrics.MemoryUsed) / float64(metrics.MemoryTotal)) * 100
		}
	case "disk":
		if metrics.DiskTotal > 0 {
			value = (float64(metrics.DiskUsed) / float64(metrics.DiskTotal)) * 100
		}
	case "load":
		value = metrics.LoadAvg
	default:
		return false, 0
	}

	// Evaluate based on operator
	triggered := false
	switch rule.Operator {
	case ">":
		triggered = value > rule.Threshold
	case "<":
		triggered = value < rule.Threshold
	case ">=":
		triggered = value >= rule.Threshold
	case "<=":
		triggered = value <= rule.Threshold
	case "==":
		triggered = value == rule.Threshold
	}

	return triggered, value
}

// sendNotifications sends alert notifications through all enabled channels
func (ae *AlertEvaluator) sendNotifications(alert *models.Alert, server *models.Server, rule *models.AlertRule) error {
	// Get all enabled notification channels
	var channels []models.NotificationChannel
	if err := ae.db.Where("enabled = ?", true).Find(&channels).Error; err != nil {
		return fmt.Errorf("failed to get notification channels: %v", err)
	}

	if len(channels) == 0 {
		log.Printf("No enabled notification channels found for alert %d", alert.ID)
		return nil
	}

	// Convert to notification channels
	notifChannels := make([]notification.NotificationChannel, len(channels))
	for i, ch := range channels {
		notifChannels[i] = notification.NotificationChannel{
			ID:      ch.ID,
			Name:    ch.Name,
			Type:    notification.ChannelType(ch.Type),
			Config:  ch.Config,
			Enabled: ch.Enabled,
		}
	}

	// Create notifier
	notifier := notification.NewNotifier(notifChannels)

	// Prepare alert notification
	notifAlert := notification.Alert{
		ID:          alert.ID,
		AlertRuleID: alert.AlertRuleID,
		ServerID:    alert.ServerID,
		ServerName:  server.Name,
		Metric:      rule.Metric,
		MetricValue: alert.MetricValue,
		Threshold:   rule.Threshold,
		Operator:    rule.Operator,
		Message:     alert.Message,
		Timestamp:   time.Now(),
	}

	// Send notification
	if err := notifier.SendNotification(notifAlert); err != nil {
		return fmt.Errorf("failed to send notification: %v", err)
	}

	log.Printf("Notifications sent successfully for alert %d", alert.ID)
	return nil
}

// CheckActiveAlerts periodically checks if active alerts should be resolved
func (ae *AlertEvaluator) CheckActiveAlerts() error {
	// Get all active alerts
	var activeAlerts []models.Alert
	if err := ae.db.Where("status = ?", "active").Find(&activeAlerts).Error; err != nil {
		return fmt.Errorf("failed to get active alerts: %v", err)
	}

	for _, alert := range activeAlerts {
		// Get the latest metrics for the server
		var metrics models.ServerMetrics
		if err := ae.db.Where("server_id = ?", alert.ServerID).Order("timestamp desc").First(&metrics).Error; err != nil {
			log.Printf("Failed to get latest metrics for server %d: %v", alert.ServerID, err)
			continue
		}

		// Get the alert rule
		var rule models.AlertRule
		if err := ae.db.First(&rule, alert.AlertRuleID).Error; err != nil {
			log.Printf("Failed to get alert rule %d: %v", alert.AlertRuleID, err)
			continue
		}

		// Check if the alert condition is still triggered
		triggered, _ := ae.evaluateRule(&rule, &metrics)
		
		// If not triggered anymore, resolve the alert
		if !triggered {
			alert.Status = "resolved"
			alert.UpdatedAt = time.Now()
			if err := ae.db.Save(&alert).Error; err != nil {
				log.Printf("Failed to update alert %d: %v", alert.ID, err)
			} else {
				log.Printf("Alert %d resolved for server %d", alert.ID, alert.ServerID)
			}
		}
	}

	return nil
}
