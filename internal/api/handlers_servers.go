package api

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/username/vps-monitor/internal/models"
	"github.com/username/vps-monitor/internal/validation"
)

// handleGetServers returns a list of all servers
func (s *Server) handleGetServers(w http.ResponseWriter, r *http.Request) {
	servers, err := s.getServers()
	if err != nil {
		log.Printf("Error retrieving servers: %v", err)
		http.Error(w, "Failed to retrieve servers", http.StatusInternalServerError)
		return
	}
	
	log.Printf("Retrieved %d servers", len(servers))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(servers)
}

// handleCreateServer creates a new server
func (s *Server) handleCreateServer(w http.ResponseWriter, r *http.Request) {
	var server models.Server
	
	if err := json.NewDecoder(r.Body).Decode(&server); err != nil {
		log.Printf("Error decoding server creation request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate server data
	validator := validation.ServerValidator{}
	if err := validator.ValidateServer(server.Name, server.IPAddress, server.Hostname, server.OS); err != nil {
		log.Printf("Server validation failed: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Generate agent key
	server.AgentKey = generateAgentKey()
	
	if err := s.createServer(&server); err != nil {
		log.Printf("Error creating server: %v", err)
		http.Error(w, "Failed to create server", http.StatusInternalServerError)
		return
	}
	
	log.Printf("Server created successfully: ID=%d, Name=%s", server.ID, server.Name)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(server)
}

// handleGetServer returns details for a specific server
func (s *Server) handleGetServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Printf("Invalid server ID provided: %v", err)
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
		return
	}
	
	server, err := s.getServerByID(serverID)
	if err != nil {
		log.Printf("Server not found with ID %d: %v", serverID, err)
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}
	
	log.Printf("Retrieved server details: ID=%d, Name=%s", server.ID, server.Name)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(server)
}

// handleUpdateServer updates a specific server
func (s *Server) handleUpdateServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Printf("Invalid server ID provided: %v", err)
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
		return
	}
	
	// Get existing server
	server, err := s.getServerByID(serverID)
	if err != nil {
		log.Printf("Server not found with ID %d: %v", serverID, err)
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}
	
	// Parse update data
	var updatedServer models.Server
	if err := json.NewDecoder(r.Body).Decode(&updatedServer); err != nil {
		log.Printf("Error decoding server update request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate server data
	validator := validation.ServerValidator{}
	if err := validator.ValidateServer(updatedServer.Name, updatedServer.IPAddress, updatedServer.Hostname, updatedServer.OS); err != nil {
		log.Printf("Server validation failed: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Update fields
	server.Name = updatedServer.Name
	server.IPAddress = updatedServer.IPAddress
	server.Hostname = updatedServer.Hostname
	server.OS = updatedServer.OS
	server.Status = updatedServer.Status
	
	if err := s.updateServer(server); err != nil {
		log.Printf("Error updating server with ID %d: %v", serverID, err)
		http.Error(w, "Failed to update server", http.StatusInternalServerError)
		return
	}
	
	log.Printf("Server updated successfully: ID=%d, Name=%s", server.ID, server.Name)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(server)
}

// handleDeleteServer deletes a specific server
func (s *Server) handleDeleteServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Printf("Invalid server ID provided: %v", err)
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
		return
	}
	
	// Check if server exists before deleting
	_, err = s.getServerByID(serverID)
	if err != nil {
		log.Printf("Server not found with ID %d: %v", serverID, err)
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}
	
	if err := s.deleteServer(serverID); err != nil {
		log.Printf("Error deleting server with ID %d: %v", serverID, err)
		http.Error(w, "Failed to delete server", http.StatusInternalServerError)
		return
	}
	
	log.Printf("Server deleted successfully: ID=%d", serverID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Server deleted successfully"})
}

// generateAgentKey generates a unique agent key
func generateAgentKey() string {
	// Generate a cryptographically secure random string
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to time-based if crypto fails (shouldn't happen)
		return "agent-key-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	return base64.URLEncoding.EncodeToString(bytes)
}