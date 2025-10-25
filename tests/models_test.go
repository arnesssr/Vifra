package tests

import (
	"testing"
	"time"

	"github.com/username/vps-monitor/internal/models"
)

func TestServerModel(t *testing.T) {
	server := models.Server{
		ID:        1,
		Name:      "test-server",
		IPAddress: "192.168.1.100",
		Hostname:  "test-server.example.com",
		OS:        "Ubuntu 22.04",
		AgentKey:  "test-agent-key",
		Status:    "active",
		LastSeen:  time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if server.Name == "" {
		t.Error("Server name should not be empty")
	}

	if server.IPAddress == "" {
		t.Error("Server IP address should not be empty")
	}

	if server.Status == "" {
		t.Error("Server status should not be empty")
	}
}

func TestServerMetricsModel(t *testing.T) {
	metrics := models.ServerMetrics{
		ID:          1,
		ServerID:    1,
		CPUUsage:    75.5,
		MemoryUsed:  8589934592,  // 8 GB
		MemoryTotal: 17179869184, // 16 GB
		DiskUsed:    107374182400,  // 100 GB
		DiskTotal:   536870912000,  // 500 GB
		NetworkIn:   1073741824, // 1 GB
		NetworkOut:  2147483648, // 2 GB
		LoadAvg:     2.5,
		Timestamp:   time.Now(),
	}

	if metrics.ServerID <= 0 {
		t.Error("Server ID should be positive")
	}

	if metrics.CPUUsage < 0 || metrics.CPUUsage > 100 {
		t.Error("CPU usage should be between 0 and 100")
	}

	if metrics.MemoryTotal == 0 {
		t.Error("Memory total should not be zero")
	}

	if metrics.MemoryUsed > metrics.MemoryTotal {
		t.Error("Memory used cannot exceed memory total")
	}

	if metrics.DiskTotal == 0 {
		t.Error("Disk total should not be zero")
	}

	if metrics.DiskUsed > metrics.DiskTotal {
		t.Error("Disk used cannot exceed disk total")
	}
}

func TestAlertRuleModel(t *testing.T) {
	serverID := 1
	rule := models.AlertRule{
		ID:        1,
		Name:      "High CPU Alert",
		ServerID:  &serverID,
		Metric:    "cpu",
		Threshold: 80.0,
		Operator:  ">",
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if rule.Name == "" {
		t.Error("Alert rule name should not be empty")
	}

	if rule.Metric == "" {
		t.Error("Alert rule metric should not be empty")
	}

	if rule.Threshold <= 0 {
		t.Error("Alert rule threshold should be positive")
	}

	if rule.Operator == "" {
		t.Error("Alert rule operator should not be empty")
	}

	validOperators := []string{">", "<", ">=", "<=", "=="}
	validOperator := false
	for _, op := range validOperators {
		if rule.Operator == op {
			validOperator = true
			break
		}
	}
	if !validOperator {
		t.Errorf("Alert rule operator '%s' is not valid", rule.Operator)
	}
}

func TestAlertModel(t *testing.T) {
	alert := models.Alert{
		ID:          1,
		AlertRuleID: 1,
		ServerID:    1,
		MetricValue: 85.5,
		Message:     "CPU usage exceeded threshold",
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if alert.AlertRuleID <= 0 {
		t.Error("Alert rule ID should be positive")
	}

	if alert.ServerID <= 0 {
		t.Error("Server ID should be positive")
	}

	if alert.Message == "" {
		t.Error("Alert message should not be empty")
	}

	if alert.Status == "" {
		t.Error("Alert status should not be empty")
	}

	validStatuses := []string{"active", "acknowledged", "resolved"}
	validStatus := false
	for _, status := range validStatuses {
		if alert.Status == status {
			validStatus = true
			break
		}
	}
	if !validStatus {
		t.Errorf("Alert status '%s' is not valid", alert.Status)
	}
}

func TestNotificationChannelModel(t *testing.T) {
	channel := models.NotificationChannel{
		ID:      1,
		Name:    "Email Alerts",
		Type:    "email",
		Config:  `{"smtp_host":"smtp.example.com","smtp_port":587}`,
		Enabled: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if channel.Name == "" {
		t.Error("Notification channel name should not be empty")
	}

	if channel.Type == "" {
		t.Error("Notification channel type should not be empty")
	}

	validTypes := []string{"email", "webhook", "slack"}
	validType := false
	for _, chanType := range validTypes {
		if channel.Type == chanType {
			validType = true
			break
		}
	}
	if !validType {
		t.Errorf("Notification channel type '%s' is not valid", channel.Type)
	}

	if channel.Config == "" {
		t.Error("Notification channel config should not be empty")
	}
}

func TestUserModel(t *testing.T) {
	user := models.User{
		ID:        1,
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "hashed_password",
		Role:      "user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if user.Username == "" {
		t.Error("Username should not be empty")
	}

	if user.Email == "" {
		t.Error("Email should not be empty")
	}

	if user.Password == "" {
		t.Error("Password should not be empty")
	}

	if user.Role == "" {
		t.Error("User role should not be empty")
	}

	validRoles := []string{"user", "admin"}
	validRole := false
	for _, role := range validRoles {
		if user.Role == role {
			validRole = true
			break
		}
	}
	if !validRole {
		t.Errorf("User role '%s' is not valid", user.Role)
	}
}
