package audit

import (
	"encoding/json"
	"log"
	"time"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	Timestamp   time.Time `json:"timestamp"`
	UserID      int       `json:"user_id,omitempty"`
	IPAddress   string    `json:"ip_address"`
	Action      string    `json:"action"`
	Resource    string    `json:"resource"`
	ResourceID  int       `json:"resource_id,omitempty"`
	Success     bool      `json:"success"`
	Description string    `json:"description,omitempty"`
	UserAgent   string    `json:"user_agent,omitempty"`
}

// Logger represents an audit logger
type Logger struct {
	// In a production environment, you might want to write to a file or external system
}

// NewLogger creates a new audit logger
func NewLogger() *Logger {
	return &Logger{}
}

// Log records an audit log entry
func (l *Logger) Log(entry AuditLog) {
	// Add timestamp if not provided
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	// Serialize to JSON for logging
	jsonEntry, err := json.Marshal(entry)
	if err != nil {
		log.Printf("Audit logging error: failed to serialize entry: %v", err)
		return
	}

	// In a production environment, you might want to:
	// 1. Write to a separate audit log file
	// 2. Send to a log aggregation system
	// 3. Store in a separate audit database
	// 4. Send to a SIEM system

	// For now, we'll log to the standard logger with an AUDIT prefix
	log.Printf("AUDIT: %s", jsonEntry)
}

// LogAction logs a security-relevant action
func (l *Logger) LogAction(userID int, ipAddress, userAgent, action, resource string, resourceID int, success bool, description string) {
	entry := AuditLog{
		UserID:      userID,
		IPAddress:   ipAddress,
		Action:      action,
		Resource:    resource,
		ResourceID:  resourceID,
		Success:     success,
		Description: description,
		UserAgent:   userAgent,
	}

	l.Log(entry)
}