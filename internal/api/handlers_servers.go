package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/username/vps-monitor/internal/models"
)

// handleGetServers returns a list of all servers
func (s *Server) handleGetServers(w http.ResponseWriter, r *http.Request) {
	servers, err := s.getServers()
	if err != nil {
		http.Error(w, "Failed to retrieve servers", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(servers)
}

// handleCreateServer creates a new server
func (s *Server) handleCreateServer(w http.ResponseWriter, r *http.Request) {
	var server models.Server
	
	if err := json.NewDecoder(r.Body).Decode(&server); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Generate agent key
	server.AgentKey = generateAgentKey()
	
	if err := s.createServer(&server); err != nil {
		http.Error(w, "Failed to create server", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(server)
}

// handleGetServer returns details for a specific server
func (s *Server) handleGetServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
		return
	}
	
	server, err := s.getServerByID(serverID)
	if err != nil {
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(server)
}

// handleUpdateServer updates a specific server
func (s *Server) handleUpdateServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
		return
	}
	
	// Get existing server
	server, err := s.getServerByID(serverID)
	if err != nil {
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}
	
	// Parse update data
	var updatedServer models.Server
	if err := json.NewDecoder(r.Body).Decode(&updatedServer); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Update fields
	server.Name = updatedServer.Name
	server.IPAddress = updatedServer.IPAddress
	server.Hostname = updatedServer.Hostname
	server.OS = updatedServer.OS
	server.Status = updatedServer.Status
	
	if err := s.updateServer(server); err != nil {
		http.Error(w, "Failed to update server", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(server)
}

// handleDeleteServer deletes a specific server
func (s *Server) handleDeleteServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
		return
	}
	
	if err := s.deleteServer(serverID); err != nil {
		http.Error(w, "Failed to delete server", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Server deleted successfully"})
}

// generateAgentKey generates a unique agent key
func generateAgentKey() string {
	// TODO: Implement proper agent key generation
	return "agent-key-" + strconv.FormatInt(time.Now().Unix(), 10)
}