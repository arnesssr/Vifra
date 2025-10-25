package api

import (
	"encoding/json"
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
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
		return
	}
	
	// Check if server exists
	_, err = s.getServerByID(serverID)
	if err != nil {
		log.Printf("Server not found with ID %d: %v", serverID, err)
		http.Error(w, "Server not found", http.StatusNotFound)
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
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
		return
	}
	
	// Check if server exists
	_, err = s.getServerByID(serverID)
	if err != nil {
		log.Printf("Server not found with ID %d: %v", serverID, err)
		http.Error(w, "Server not found", http.StatusNotFound)
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
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(metrics)
}

// handlePostMetrics handles metrics submission from agents
func (s *Server) handlePostMetrics(w http.ResponseWriter, r *http.Request) {
	var metrics models.ServerMetrics
	
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		log.Printf("Error decoding metrics submission request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate metrics data
	validator := validation.MetricsValidator{}
	if err := validator.ValidateMetricsData(metrics.ServerID, metrics.CPUUsage, metrics.MemoryUsed, metrics.MemoryTotal, metrics.DiskUsed, metrics.DiskTotal); err != nil {
		log.Printf("Metrics validation failed: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Check if server exists
	_, err := s.getServerByID(metrics.ServerID)
	if err != nil {
		log.Printf("Server not found with ID %d: %v", metrics.ServerID, err)
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
		http.Error(w, "Failed to save metrics", http.StatusInternalServerError)
		return
	}
	
	log.Printf("Metrics saved successfully for server ID %d", metrics.ServerID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(metrics)
}