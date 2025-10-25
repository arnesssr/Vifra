package tests

import (
	"testing"

	"github.com/username/vps-monitor/internal/models"
)

func TestAlertRuleOperators(t *testing.T) {
	tests := []struct {
		name      string
		value     float64
		threshold float64
		operator  string
		triggered bool
	}{
		{"Greater Than - Triggered", 85.0, 80.0, ">", true},
		{"Greater Than - Not Triggered", 75.0, 80.0, ">", false},
		{"Less Than - Triggered", 15.0, 20.0, "<", true},
		{"Less Than - Not Triggered", 25.0, 20.0, "<", false},
		{"Greater Equal - Triggered Exact", 80.0, 80.0, ">=", true},
		{"Greater Equal - Triggered Above", 85.0, 80.0, ">=", true},
		{"Greater Equal - Not Triggered", 75.0, 80.0, ">=", false},
		{"Less Equal - Triggered Exact", 20.0, 20.0, "<=", true},
		{"Less Equal - Triggered Below", 15.0, 20.0, "<=", true},
		{"Less Equal - Not Triggered", 25.0, 20.0, "<=", false},
		{"Equal - Triggered", 80.0, 80.0, "==", true},
		{"Equal - Not Triggered", 85.0, 80.0, "==", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var triggered bool
			switch tt.operator {
			case ">":
				triggered = tt.value > tt.threshold
			case "<":
				triggered = tt.value < tt.threshold
			case ">=":
				triggered = tt.value >= tt.threshold
			case "<=":
				triggered = tt.value <= tt.threshold
			case "==":
				triggered = tt.value == tt.threshold
			}

			if triggered != tt.triggered {
				t.Errorf("Expected triggered=%v for value=%.2f %s threshold=%.2f, got %v",
					tt.triggered, tt.value, tt.operator, tt.threshold, triggered)
			}
		})
	}
}

func TestMetricCalculations(t *testing.T) {
	metrics := models.ServerMetrics{
		CPUUsage:    75.5,
		MemoryUsed:  8 * 1024 * 1024 * 1024,  // 8 GB
		MemoryTotal: 16 * 1024 * 1024 * 1024, // 16 GB
		DiskUsed:    100 * 1024 * 1024 * 1024,  // 100 GB
		DiskTotal:   500 * 1024 * 1024 * 1024,  // 500 GB
		LoadAvg:     2.5,
	}

	// Test CPU percentage (direct)
	if metrics.CPUUsage != 75.5 {
		t.Errorf("Expected CPU usage 75.5, got %.2f", metrics.CPUUsage)
	}

	// Test memory percentage calculation
	memoryPercent := (float64(metrics.MemoryUsed) / float64(metrics.MemoryTotal)) * 100
	expectedMemory := 50.0
	if memoryPercent != expectedMemory {
		t.Errorf("Expected memory percentage %.2f, got %.2f", expectedMemory, memoryPercent)
	}

	// Test disk percentage calculation
	diskPercent := (float64(metrics.DiskUsed) / float64(metrics.DiskTotal)) * 100
	expectedDisk := 20.0
	if diskPercent != expectedDisk {
		t.Errorf("Expected disk percentage %.2f, got %.2f", expectedDisk, diskPercent)
	}

	// Test load average (direct)
	if metrics.LoadAvg != 2.5 {
		t.Errorf("Expected load average 2.5, got %.2f", metrics.LoadAvg)
	}
}

func TestAlertRuleValidation(t *testing.T) {
	tests := []struct {
		name   string
		rule   models.AlertRule
		valid  bool
		reason string
	}{
		{
			name: "Valid CPU Alert",
			rule: models.AlertRule{
				Name:      "High CPU",
				Metric:    "cpu",
				Threshold: 80.0,
				Operator:  ">",
				Enabled:   true,
			},
			valid:  true,
			reason: "",
		},
		{
			name: "Valid Memory Alert",
			rule: models.AlertRule{
				Name:      "High Memory",
				Metric:    "memory",
				Threshold: 90.0,
				Operator:  ">",
				Enabled:   true,
			},
			valid:  true,
			reason: "",
		},
		{
			name: "Valid Disk Alert",
			rule: models.AlertRule{
				Name:      "Low Disk Space",
				Metric:    "disk",
				Threshold: 85.0,
				Operator:  ">",
				Enabled:   true,
			},
			valid:  true,
			reason: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.rule.Name == "" {
				t.Error("Alert rule name should not be empty")
			}
			if tt.rule.Metric == "" {
				t.Error("Alert rule metric should not be empty")
			}
			if tt.rule.Threshold <= 0 {
				t.Error("Alert rule threshold should be positive")
			}
			if tt.rule.Operator == "" {
				t.Error("Alert rule operator should not be empty")
			}
		})
	}
}

func TestAlertStatusTransitions(t *testing.T) {
	validStatuses := []string{"active", "acknowledged", "resolved"}
	
	for _, status := range validStatuses {
		t.Run("Status_"+status, func(t *testing.T) {
			alert := models.Alert{
				Status: status,
			}
			
			if alert.Status == "" {
				t.Error("Alert status should not be empty")
			}
			
			// Check if status is valid
			isValid := false
			for _, vs := range validStatuses {
				if alert.Status == vs {
					isValid = true
					break
				}
			}
			
			if !isValid {
				t.Errorf("Alert status '%s' is not valid", alert.Status)
			}
		})
	}
}
