package api

import (
	"crypto/rand"
	"encoding/base64"
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

// handleGetServers returns a list of all servers
func (s *Server) handleGetServers(w http.ResponseWriter, r *http.Request) {
	servers, err := s.getServers()
	if err != nil {
		log.Printf("Error retrieving servers: %v", err)
		
		// Audit log failed server retrieval attempt
		s.auditLogger.LogAction(
			s.getUserIDFromContext(r),
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"get_servers",
			"servers",
			0,
			false,
			"Database error: "+err.Error(),
		)
		
		http.Error(w, "Failed to retrieve servers", http.StatusInternalServerError)
		return
	}
	
	log.Printf("Retrieved %d servers", len(servers))
	
	// Audit log successful server retrieval
	s.auditLogger.LogAction(
		s.getUserIDFromContext(r),
		s.getIPAddress(r),
		r.Header.Get("User-Agent"),
		"get_servers",
		"servers",
		0,
		true,
		fmt.Sprintf("Retrieved %d servers", len(servers)),
	)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(servers)
}

// handleCreateServer creates a new server
func (s *Server) handleCreateServer(w http.ResponseWriter, r *http.Request) {
	var server models.Server
	
	if err := json.NewDecoder(r.Body).Decode(&server); err != nil {
		log.Printf("Error decoding server creation request: %v", err)
		
		// Audit log failed server creation attempt
		s.auditLogger.LogAction(
			s.getUserIDFromContext(r),
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"create_server",
			"servers",
			0,
			false,
			"Invalid request body: "+err.Error(),
		)
		
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate server data
	validator := validation.ServerValidator{}
	if err := validator.ValidateServer(server.Name, server.IPAddress, server.Hostname, server.OS); err != nil {
		log.Printf("Server validation failed: %v", err)
		
		// Audit log failed server creation attempt
		s.auditLogger.LogAction(
			s.getUserIDFromContext(r),
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"create_server",
			"servers",
			0,
			false,
			"Validation failed: "+err.Error(),
		)
		
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Generate agent key
	agentKey, err := s.generateAgentKey()
	if err != nil {
		log.Printf("Error generating agent key: %v", err)
		
		// Audit log failed server creation attempt
		s.auditLogger.LogAction(
			s.getUserIDFromContext(r),
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"create_server",
			"servers",
			0,
			false,
			"Failed to generate agent key: "+err.Error(),
		)
		
		http.Error(w, "Failed to generate agent key", http.StatusInternalServerError)
		return
	}
	server.AgentKey = agentKey
	
	if err := s.createServer(&server); err != nil {
		log.Printf("Error creating server: %v", err)
		
		// Audit log failed server creation attempt
		s.auditLogger.LogAction(
			s.getUserIDFromContext(r),
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"create_server",
			"servers",
			0,
			false,
			"Database error: "+err.Error(),
		)
		
		http.Error(w, "Failed to create server", http.StatusInternalServerError)
		return
	}
	
	log.Printf("Server created successfully: ID=%d, Name=%s", server.ID, server.Name)
	
	// Audit log successful server creation
	s.auditLogger.LogAction(
		s.getUserIDFromContext(r),
		s.getIPAddress(r),
		r.Header.Get("User-Agent"),
		"create_server",
		"servers",
		server.ID,
		true,
		"Server created successfully",
	)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(server)
}

// handleGetServer returns details for a specific server
func (s *Server) handleGetServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Printf("Invalid server ID provided: %v", err)
		
		// Audit log failed server retrieval attempt
		s.auditLogger.LogAction(
			s.getUserIDFromContext(r),
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"get_server",
			"servers",
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
		
		// Audit log failed server retrieval attempt
		s.auditLogger.LogAction(
			s.getUserIDFromContext(r),
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"get_server",
			"servers",
			serverID,
			false,
			"Access denied: "+err.Error(),
		)
		
		// Return a generic not found error to prevent enumeration
		s.errorHandler.HandleNotFound(w, r, fmt.Errorf("server not found"), "Server not found")
		return
	}
	
	server, err := s.getServerByID(serverID)
	if err != nil {
		log.Printf("Server not found with ID %d: %v", serverID, err)
		
		// Audit log failed server retrieval attempt
		s.auditLogger.LogAction(
			s.getUserIDFromContext(r),
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"get_server",
			"servers",
			serverID,
			false,
			"Server not found: "+err.Error(),
		)
		
		s.errorHandler.HandleNotFound(w, r, err, "Server not found")
		return
	}
	
	log.Printf("Retrieved server details: ID=%d, Name=%s", server.ID, server.Name)
	
	// Audit log successful server retrieval
	s.auditLogger.LogAction(
		s.getUserIDFromContext(r),
		s.getIPAddress(r),
		r.Header.Get("User-Agent"),
		"get_server",
		"servers",
		server.ID,
		true,
		"Server details retrieved successfully",
	)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(server)
}

// handleUpdateServer updates a specific server
func (s *Server) handleUpdateServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Printf("Invalid server ID provided: %v", err)
		
		// Audit log failed server update attempt
		s.auditLogger.LogAction(
			s.getUserIDFromContext(r),
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"update_server",
			"servers",
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
		
		// Audit log failed server update attempt
		s.auditLogger.LogAction(
			s.getUserIDFromContext(r),
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"update_server",
			"servers",
			serverID,
			false,
			"Access denied: "+err.Error(),
		)
		
		// Return a generic not found error to prevent enumeration
		s.errorHandler.HandleNotFound(w, r, fmt.Errorf("server not found"), "Server not found")
		return
	}
	
	// Get existing server
	server, err := s.getServerByID(serverID)
	if err != nil {
		log.Printf("Server not found with ID %d: %v", serverID, err)
		
		// Audit log failed server update attempt
		s.auditLogger.LogAction(
			s.getUserIDFromContext(r),
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"update_server",
			"servers",
			serverID,
			false,
			"Server not found: "+err.Error(),
		)
		
		s.errorHandler.HandleNotFound(w, r, err, "Server not found")
		return
	}
	
	// Parse update data
	var updatedServer models.Server
	if err := json.NewDecoder(r.Body).Decode(&updatedServer); err != nil {
		log.Printf("Error decoding server update request: %v", err)
		
		// Audit log failed server update attempt
		s.auditLogger.LogAction(
			s.getUserIDFromContext(r),
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"update_server",
			"servers",
			serverID,
			false,
			"Invalid request body: "+err.Error(),
		)
		
		s.errorHandler.HandleBadRequest(w, r, err, "Invalid request body")
		return
	}
	
	// Validate server data
	validator := validation.ServerValidator{}
	if err := validator.ValidateServer(updatedServer.Name, updatedServer.IPAddress, updatedServer.Hostname, updatedServer.OS); err != nil {
		log.Printf("Server validation failed: %v", err)
		
		// Audit log failed server update attempt
		s.auditLogger.LogAction(
			s.getUserIDFromContext(r),
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"update_server",
			"servers",
			serverID,
			false,
			"Validation failed: "+err.Error(),
		)
		
		s.errorHandler.HandleBadRequest(w, r, err, "Validation failed")
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
		
		// Audit log failed server update attempt
		s.auditLogger.LogAction(
			s.getUserIDFromContext(r),
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"update_server",
			"servers",
			serverID,
			false,
			"Database error: "+err.Error(),
		)
		
		s.errorHandler.HandleInternalServerError(w, r, err, "Failed to update server")
		return
	}
	
	log.Printf("Server updated successfully: ID=%d, Name=%s", server.ID, server.Name)
	
	// Audit log successful server update
		s.auditLogger.LogAction(
		s.getUserIDFromContext(r),
		s.getIPAddress(r),
		r.Header.Get("User-Agent"),
		"update_server",
		"servers",
		server.ID,
		true,
		"Server updated successfully",
	)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(server)
}

// handleDeleteServer deletes a specific server
func (s *Server) handleDeleteServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Printf("Invalid server ID provided: %v", err)
		
		// Audit log failed server deletion attempt
		s.auditLogger.LogAction(
			s.getUserIDFromContext(r),
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"delete_server",
			"servers",
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
		
		// Audit log failed server deletion attempt
		s.auditLogger.LogAction(
			s.getUserIDFromContext(r),
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"delete_server",
			"servers",
			serverID,
			false,
			"Access denied: "+err.Error(),
		)
		
		// Return a generic not found error to prevent enumeration
		s.errorHandler.HandleNotFound(w, r, fmt.Errorf("server not found"), "Server not found")
		return
	}
	
	// Check if server exists before deleting
	_, err = s.getServerByID(serverID)
	if err != nil {
		log.Printf("Server not found with ID %d: %v", serverID, err)
		
		// Audit log failed server deletion attempt
		s.auditLogger.LogAction(
			s.getUserIDFromContext(r),
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"delete_server",
			"servers",
			serverID,
			false,
			"Server not found: "+err.Error(),
		)
		
		s.errorHandler.HandleNotFound(w, r, err, "Server not found")
		return
	}
	
	if err := s.deleteServer(serverID); err != nil {
		log.Printf("Error deleting server with ID %d: %v", serverID, err)
		
		// Audit log failed server deletion attempt
		s.auditLogger.LogAction(
			s.getUserIDFromContext(r),
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"delete_server",
			"servers",
			serverID,
			false,
			"Database error: "+err.Error(),
		)
		
		s.errorHandler.HandleInternalServerError(w, r, err, "Failed to delete server")
		return
	}
	
	log.Printf("Server deleted successfully: ID=%d", serverID)
	
	// Audit log successful server deletion
	s.auditLogger.LogAction(
		s.getUserIDFromContext(r),
		s.getIPAddress(r),
		r.Header.Get("User-Agent"),
		"delete_server",
		"servers",
		serverID,
		true,
		"Server deleted successfully",
	)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Server deleted successfully"})
}

// generateAgentKey generates a unique agent key
func (s *Server) generateAgentKey() (string, error) {
	// Generate a cryptographically secure random string
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to time-based if crypto fails (shouldn't happen)
		key := "agent-key-" + strconv.FormatInt(time.Now().UnixNano(), 10)
		if s.encryptor != nil {
			encryptedKey, err := s.encryptor.Encrypt([]byte(key))
			if err != nil {
				return key, nil // Return unencrypted if encryption fails
			}
			return encryptedKey, nil
		}
		return key, nil
	}
	key := base64.URLEncoding.EncodeToString(bytes)
	
	// Encrypt the key if encryptor is available
	if s.encryptor != nil {
		encryptedKey, err := s.encryptor.Encrypt([]byte(key))
		if err != nil {
			return key, nil // Return unencrypted if encryption fails
		}
		return encryptedKey, nil
	}
	
	return key, nil
}