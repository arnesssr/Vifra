package models

import (
	"time"
)

// Server represents a monitored VPS server
type Server struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	IPAddress   string    `json:"ip_address" gorm:"not null"`
	Hostname    string    `json:"hostname"`
	OS          string    `json:"os"`
	AgentKey    string    `json:"agent_key" gorm:"uniqueIndex"`
	Status      string    `json:"status" gorm:"default:'active'"`
	LastSeen    time.Time `json:"last_seen"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ServerMetrics represents the metrics collected from a server
type ServerMetrics struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	ServerID  int       `json:"server_id" gorm:"index"`
	CPUUsage  float64   `json:"cpu_usage"`
	MemoryUsed uint64   `json:"memory_used"`
	MemoryTotal uint64  `json:"memory_total"`
	DiskUsed  uint64    `json:"disk_used"`
	DiskTotal uint64    `json:"disk_total"`
	NetworkIn uint64    `json:"network_in"`
	NetworkOut uint64   `json:"network_out"`
	LoadAvg   float64   `json:"load_avg"`
	Timestamp time.Time `json:"timestamp" gorm:"index"`
}

// AlertRule represents a rule for triggering alerts
type AlertRule struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	ServerID    *int      `json:"server_id"` // Null for global rules
	Metric      string    `json:"metric" gorm:"not null"` // cpu, memory, disk, network
	Threshold   float64   `json:"threshold" gorm:"not null"`
	Operator    string    `json:"operator" gorm:"not null"` // >, <, >=, <=, ==
	Enabled     bool      `json:"enabled" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Alert represents a triggered alert
type Alert struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	AlertRuleID int       `json:"alert_rule_id" gorm:"not null"`
	ServerID    int       `json:"server_id" gorm:"not null"`
	MetricValue float64   `json:"metric_value"`
	Message     string    `json:"message"`
	Status      string    `json:"status" gorm:"default:'active'"` // active, acknowledged, resolved
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// User represents a system user
type User struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"uniqueIndex;not null"`
	Email     string    `json:"email" gorm:"uniqueIndex;not null"`
	Password  string    `json:"-" gorm:"not null"` // Never serialize password
	Role      string    `json:"role" gorm:"default:'user'"` // user, admin
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}