package validation

import (
	"net"
	"regexp"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

// ValidationErrors represents a collection of validation errors
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return ""
	}
	return ve[0].Message
}

// ServerValidator validates server data
type ServerValidator struct{}

// ValidateName validates server name
func (sv ServerValidator) ValidateName(name string) error {
	if name == "" {
		return ValidationError{Field: "name", Message: "Server name is required"}
	}
	if len(name) > 100 {
		return ValidationError{Field: "name", Message: "Server name must be less than 100 characters"}
	}
	return nil
}

// ValidateIPAddress validates IP address
func (sv ServerValidator) ValidateIPAddress(ip string) error {
	if ip == "" {
		return ValidationError{Field: "ip_address", Message: "IP address is required"}
	}
	
	// Check if it's a valid IP address
	if net.ParseIP(ip) == nil {
		return ValidationError{Field: "ip_address", Message: "Invalid IP address format"}
	}
	
	return nil
}

// ValidateHostname validates hostname
func (sv ServerValidator) ValidateHostname(hostname string) error {
	if hostname == "" {
		return nil // Hostname is optional
	}
	
	if len(hostname) > 255 {
		return ValidationError{Field: "hostname", Message: "Hostname must be less than 255 characters"}
	}
	
	// Basic hostname validation regex
	hostnameRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
	if !hostnameRegex.MatchString(hostname) {
		return ValidationError{Field: "hostname", Message: "Invalid hostname format"}
	}
	
	return nil
}

// ValidateOS validates OS field
func (sv ServerValidator) ValidateOS(os string) error {
	if os == "" {
		return nil // OS is optional
	}
	
	if len(os) > 100 {
		return ValidationError{Field: "os", Message: "OS must be less than 100 characters"}
	}
	
	return nil
}

// ValidateServer validates all server fields
func (sv ServerValidator) ValidateServer(name, ip, hostname, os string) error {
	var errors ValidationErrors
	
	if err := sv.ValidateName(name); err != nil {
		errors = append(errors, err.(ValidationError))
	}
	
	if err := sv.ValidateIPAddress(ip); err != nil {
		errors = append(errors, err.(ValidationError))
	}
	
	if err := sv.ValidateHostname(hostname); err != nil {
		if _, ok := err.(ValidationError); ok {
			errors = append(errors, err.(ValidationError))
		}
	}
	
	if err := sv.ValidateOS(os); err != nil {
		if _, ok := err.(ValidationError); ok {
			errors = append(errors, err.(ValidationError))
		}
	}
	
	if len(errors) > 0 {
		return errors
	}
	
	return nil
}

// MetricsValidator validates metrics data
type MetricsValidator struct{}

// ValidateMetricsData validates server metrics data
func (mv MetricsValidator) ValidateMetricsData(serverID int, cpuUsage float64, memoryUsed, memoryTotal, diskUsed, diskTotal uint64) error {
	var errors ValidationErrors
	
	if serverID == 0 {
		errors = append(errors, ValidationError{Field: "server_id", Message: "Server ID is required"})
	}
	
	if cpuUsage < 0 || cpuUsage > 100 {
		errors = append(errors, ValidationError{Field: "cpu_usage", Message: "CPU usage must be between 0 and 100"})
	}
	
	if memoryTotal > 0 && memoryUsed > memoryTotal {
		errors = append(errors, ValidationError{Field: "memory_used", Message: "Memory used cannot exceed memory total"})
	}
	
	if diskTotal > 0 && diskUsed > diskTotal {
		errors = append(errors, ValidationError{Field: "disk_used", Message: "Disk used cannot exceed disk total"})
	}
	
	if len(errors) > 0 {
		return errors
	}
	
	return nil
}

// AlertValidator validates alert data
type AlertValidator struct{}

// ValidateAlertData validates alert data
func (av AlertValidator) ValidateAlertData(alertRuleID, serverID int, status string) error {
	var errors ValidationErrors
	
	if alertRuleID == 0 {
		errors = append(errors, ValidationError{Field: "alert_rule_id", Message: "Alert rule ID is required"})
	}
	
	if serverID == 0 {
		errors = append(errors, ValidationError{Field: "server_id", Message: "Server ID is required"})
	}
	
	// Validate status if provided
	if status != "" {
		validStatuses := map[string]bool{
			"active":       true,
			"acknowledged": true,
			"resolved":     true,
		}
		if !validStatuses[status] {
			errors = append(errors, ValidationError{Field: "status", Message: "Invalid status value"})
		}
	}
	
	if len(errors) > 0 {
		return errors
	}
	
	return nil
}