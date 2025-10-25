package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/username/vps-monitor/internal/models"
	"github.com/username/vps-monitor/internal/validation"
)

// handleGetMetrics returns real-time metrics for a server
func (s *Server) handleGetMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Printf("Invalid server ID provided: %v", err)
		
		// Audit log failed metrics retrieval attempt
		s.auditLogger.LogAction(
			s.getUserIDFromContext(r),
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"get_metrics",
			"metrics",
			0,
			false,
			"Invalid server ID: "+err.Error(),
		)
		
		s.errorHandler.HandleBadRequest(w, r, err, "Invalid server ID")
		return
	}
	
	// Check if user has access to this server to prevent enumeration
	if err := s.checkServerAccess(serverID); err != nil {
		log.Printf("Access denied to server ID %d: %v", serverID, err)
		
		// Audit log failed metrics retrieval attempt
		s.auditLogger.LogAction(
			s.getUserIDFromContext(r),
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"get_metrics",
			"metrics",
			serverID,
			false,
			"Access denied: "+err.Error(),
		)
		
		// Return a generic not found error to prevent enumeration
		s.errorHandler.HandleNotFound(w, r, fmt.Errorf("server not found"), "Server not found")
		return
	}
	
	// Check if server exists
	_, err = s.getServerByID(serverID)
	if err != nil {
		log.Printf("Server not found with ID %d: %v", serverID, err)
		
		// Audit log failed metrics retrieval attempt
		s.auditLogger.LogAction(
			s.getUserIDFromContext(r),
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"get_metrics",
			"metrics",
			serverID,
			false,
			"Server not found: "+err.Error(),
		)
		
		s.errorHandler.HandleNotFound(w, r, err, "Server not found")
		return
	}
	
	// Get latest metrics for the server
	metrics, err := s.getMetrics(serverID)
	if err != nil {
		// This is not an error condition, just no metrics yet
		log.Printf("No metrics found for server ID %d, returning empty metrics", serverID)
		metrics = &models.ServerMetrics{
			ServerID: serverID,
		}
	} else {
		log.Printf("Retrieved metrics for server ID %d", serverID)
	}
	
	// Audit log successful metrics retrieval
	s.auditLogger.LogAction(
		s.getUserIDFromContext(r),
		s.getIPAddress(r),
		r.Header.Get("User-Agent"),
		"get_metrics",
		"metrics",
		serverID,
		true,
		"Metrics retrieved successfully",
	)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(metrics)
}

// handleGetMetricsHistory returns historical metrics for a server
func (s *Server) handleGetMetricsHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Printf("Invalid server ID provided: %v", err)
		
		// Audit log failed metrics history retrieval attempt
		s.auditLogger.LogAction(
			s.getUserIDFromContext(r),
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"get_metrics_history",
			"metrics",
			0,
			false,
			"Invalid server ID: "+err.Error(),
		)
		
		s.errorHandler.HandleBadRequest(w, r, err, "Invalid server ID")
		return
	}
	
	// Check if user has access to this server to prevent enumeration
	if err := s.checkServerAccess(serverID); err != nil {
		log.Printf("Access denied to server ID %d: %v", serverID, err)
		
		// Audit log failed metrics history retrieval attempt
		s.auditLogger.LogAction(
			s.getUserIDFromContext(r),
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"get_metrics_history",
			"metrics",
			serverID,
			false,
			"Access denied: "+err.Error(),
		)
		
		// Return a generic not found error to prevent enumeration
		s.errorHandler.HandleNotFound(w, r, fmt.Errorf("server not found"), "Server not found")
		return
	}
	
	// Check if server exists
	_, err = s.getServerByID(serverID)
	if err != nil {
		log.Printf("Server not found with ID %d: %v", serverID, err)
		
		// Audit log failed metrics history retrieval attempt
		s.auditLogger.LogAction(
			s.getUserIDFromContext(r),
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"get_metrics_history",
			"metrics",
			serverID,
			false,
			"Server not found: "+err.Error(),
		)
		
		s.errorHandler.HandleNotFound(w, r, err, "Server not found")
		return
	}
	
	// Get metrics history for the server
	metrics, err := s.getMetricsHistory(serverID)
	if err != nil {
		// This is not an error condition, just no metrics history yet
		log.Printf("No metrics history found for server ID %d, returning empty array", serverID)
		metrics = []models.ServerMetrics{}
	} else {
		log.Printf("Retrieved %d metrics history records for server ID %d", len(metrics), serverID)
	}
	
	// Audit log successful metrics history retrieval
	s.auditLogger.LogAction(
		s.getUserIDFromContext(r),
		s.getIPAddress(r),
		r.Header.Get("User-Agent"),
		"get_metrics_history",
		"metrics",
		serverID,
		true,
		fmt.Sprintf("Retrieved %d metrics history records", len(metrics)),
	)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(metrics)
}

// handlePostMetrics handles metrics submission from agents
func (s *Server) handlePostMetrics(w http.ResponseWriter, r *http.Request) {
	var metrics models.ServerMetrics
	
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		log.Printf("Error decoding metrics submission request: %v", err)
		
		// Audit log failed metrics submission attempt
		s.auditLogger.LogAction(
			0, // No user ID for agent requests
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"submit_metrics",
			"metrics",
			0,
			false,
			"Invalid request body: "+err.Error(),
		)
		
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate metrics data
	validator := validation.MetricsValidator{}
	if err := validator.ValidateMetricsData(metrics.ServerID, metrics.CPUUsage, metrics.MemoryUsed, metrics.MemoryTotal, metrics.DiskUsed, metrics.DiskTotal); err != nil {
		log.Printf("Metrics validation failed: %v", err)
		
		// Audit log failed metrics submission attempt
		s.auditLogger.LogAction(
			0, // No user ID for agent requests
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"submit_metrics",
			"metrics",
			metrics.ServerID,
			false,
			"Validation failed: "+err.Error(),
		)
		
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Check if server exists
	_, err := s.getServerByID(metrics.ServerID)
	if err != nil {
		log.Printf("Server not found with ID %d: %v", metrics.ServerID, err)
		
		// Audit log failed metrics submission attempt
		s.auditLogger.LogAction(
			0, // No user ID for agent requests
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"submit_metrics",
			"metrics",
			metrics.ServerID,
			false,
			"Server not found: "+err.Error(),
		)
		
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}
	
	// Set timestamp if not provided
	if metrics.Timestamp.IsZero() {
		metrics.Timestamp = time.Now()
	}
	
	// Save metrics
	if err := s.createMetrics(&metrics); err != nil {
		log.Printf("Error saving metrics for server ID %d: %v", metrics.ServerID, err)
		
		// Audit log failed metrics submission attempt
		s.auditLogger.LogAction(
			0, // No user ID for agent requests
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"submit_metrics",
			"metrics",
			metrics.ServerID,
			false,
			"Database error: "+err.Error(),
		)
		
		http.Error(w, "Failed to save metrics", http.StatusInternalServerError)
		return
	}
	
	log.Printf("Metrics saved successfully for server ID %d", metrics.ServerID)
	
	// Evaluate alert rules for the submitted metrics
	if err := s.alertEvaluator.EvaluateMetrics(&metrics); err != nil {
		// Log the error but don't fail the request
		log.Printf("Failed to evaluate alert rules for server ID %d: %v", metrics.ServerID, err)
	}
	
	// Audit log successful metrics submission
	s.auditLogger.LogAction(
		0, // No user ID for agent requests
		s.getIPAddress(r),
		r.Header.Get("User-Agent"),
		"submit_metrics",
		"metrics",
		metrics.ServerID,
		true,
		"Metrics saved successfully",
	)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(metrics)
}
