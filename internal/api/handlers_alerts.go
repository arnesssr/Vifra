package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/username/vps-monitor/internal/models"
	"github.com/username/vps-monitor/internal/validation"
)

// handleGetAlerts returns a list of alerts
func (s *Server) handleGetAlerts(w http.ResponseWriter, r *http.Request) {
	alerts, err := s.getAlerts()
	if err != nil {
		log.Printf("Error retrieving alerts: %v", err)
		http.Error(w, "Failed to retrieve alerts", http.StatusInternalServerError)
		return
	}
	
	log.Printf("Retrieved %d alerts", len(alerts))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(alerts)
}

// handleCreateAlert creates a new alert
func (s *Server) handleCreateAlert(w http.ResponseWriter, r *http.Request) {
	var alert models.Alert
	
	if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
		log.Printf("Error decoding alert creation request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate alert data
	validator := validation.AlertValidator{}
	if err := validator.ValidateAlertData(alert.AlertRuleID, alert.ServerID, alert.Status); err != nil {
		log.Printf("Alert validation failed: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Check if alert rule exists
	_, err := s.getAlertRuleByID(alert.AlertRuleID)
	if err != nil {
		log.Printf("Alert rule not found with ID %d: %v", alert.AlertRuleID, err)
		http.Error(w, "Alert rule not found", http.StatusNotFound)
		return
	}
	
	// Check if server exists
	_, err = s.getServerByID(alert.ServerID)
	if err != nil {
		log.Printf("Server not found with ID %d: %v", alert.ServerID, err)
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}
	
	// Set default status if not provided
	if alert.Status == "" {
		alert.Status = "active"
	}
	
	if err := s.createAlert(&alert); err != nil {
		log.Printf("Error creating alert: %v", err)
		http.Error(w, "Failed to create alert", http.StatusInternalServerError)
		return
	}
	
	log.Printf("Alert created successfully: ID=%d, RuleID=%d, ServerID=%d", alert.ID, alert.AlertRuleID, alert.ServerID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(alert)
}

// handleGetAlert returns details for a specific alert
func (s *Server) handleGetAlert(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	alertID, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Printf("Invalid alert ID provided: %v", err)
		http.Error(w, "Invalid alert ID", http.StatusBadRequest)
		return
	}
	
	alert, err := s.getAlertByID(alertID)
	if err != nil {
		log.Printf("Alert not found with ID %d: %v", alertID, err)
		http.Error(w, "Alert not found", http.StatusNotFound)
		return
	}
	
	log.Printf("Retrieved alert details: ID=%d, RuleID=%d, ServerID=%d", alert.ID, alert.AlertRuleID, alert.ServerID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(alert)
}

// handleUpdateAlert updates a specific alert
func (s *Server) handleUpdateAlert(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	alertID, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Printf("Invalid alert ID provided: %v", err)
		http.Error(w, "Invalid alert ID", http.StatusBadRequest)
		return
	}
	
	// Get existing alert
	alert, err := s.getAlertByID(alertID)
	if err != nil {
		log.Printf("Alert not found with ID %d: %v", alertID, err)
		http.Error(w, "Alert not found", http.StatusNotFound)
		return
	}
	
	// Parse update data
	var updatedAlert models.Alert
	if err := json.NewDecoder(r.Body).Decode(&updatedAlert); err != nil {
		log.Printf("Error decoding alert update request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate alert data (only validate provided fields)
	tempAlert := *alert
	if updatedAlert.AlertRuleID != 0 {
		tempAlert.AlertRuleID = updatedAlert.AlertRuleID
	}
	if updatedAlert.ServerID != 0 {
		tempAlert.ServerID = updatedAlert.ServerID
	}
	if updatedAlert.Status != "" {
		tempAlert.Status = updatedAlert.Status
	}
	
	validator := validation.AlertValidator{}
	if err := validator.ValidateAlertData(tempAlert.AlertRuleID, tempAlert.ServerID, tempAlert.Status); err != nil {
		log.Printf("Alert validation failed: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Update fields
	if updatedAlert.AlertRuleID != 0 {
		alert.AlertRuleID = updatedAlert.AlertRuleID
	}
	if updatedAlert.ServerID != 0 {
		alert.ServerID = updatedAlert.ServerID
	}
	if updatedAlert.MetricValue != 0 {
		alert.MetricValue = updatedAlert.MetricValue
	}
	if updatedAlert.Message != "" {
		alert.Message = updatedAlert.Message
	}
	if updatedAlert.Status != "" {
		alert.Status = updatedAlert.Status
	}
	
	if err := s.updateAlert(alert); err != nil {
		log.Printf("Error updating alert with ID %d: %v", alertID, err)
		http.Error(w, "Failed to update alert", http.StatusInternalServerError)
		return
	}
	
	log.Printf("Alert updated successfully: ID=%d, RuleID=%d, ServerID=%d", alert.ID, alert.AlertRuleID, alert.ServerID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(alert)
}

// handleDeleteAlert deletes a specific alert
func (s *Server) handleDeleteAlert(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	alertID, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Printf("Invalid alert ID provided: %v", err)
		http.Error(w, "Invalid alert ID", http.StatusBadRequest)
		return
	}
	
	// Check if alert exists before deleting
	_, err = s.getAlertByID(alertID)
	if err != nil {
		log.Printf("Alert not found with ID %d: %v", alertID, err)
		http.Error(w, "Alert not found", http.StatusNotFound)
		return
	}
	
	if err := s.deleteAlert(alertID); err != nil {
		log.Printf("Error deleting alert with ID %d: %v", alertID, err)
		http.Error(w, "Failed to delete alert", http.StatusInternalServerError)
		return
	}
	
	log.Printf("Alert deleted successfully: ID=%d", alertID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Alert deleted successfully"})
}