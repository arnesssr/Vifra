package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/username/vps-monitor/internal/models"
)

// handleGetNotificationChannels returns a list of notification channels
func (s *Server) handleGetNotificationChannels(w http.ResponseWriter, r *http.Request) {
	channels, err := s.getNotificationChannels()
	if err != nil {
		log.Printf("Error retrieving notification channels: %v", err)
		
		// Audit log failed notification channels retrieval attempt
		s.auditLogger.LogAction(
			s.getUserIDFromContext(r),
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"get_notification_channels",
			"notification_channels",
			0,
			false,
			"Database error: "+err.Error(),
		)
		
		s.errorHandler.HandleInternalServerError(w, r, err, "Failed to retrieve notification channels")
		return
	}
	
	log.Printf("Retrieved %d notification channels", len(channels))
	
	// Audit log successful notification channels retrieval
	s.auditLogger.LogAction(
		s.getUserIDFromContext(r),
		s.getIPAddress(r),
		r.Header.Get("User-Agent"),
		"get_notification_channels",
		"notification_channels",
		0,
		true,
		fmt.Sprintf("Retrieved %d notification channels", len(channels)),
	)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(channels)
}

// handleCreateNotificationChannel creates a new notification channel
func (s *Server) handleCreateNotificationChannel(w http.ResponseWriter, r *http.Request) {
	var channel models.NotificationChannel
	
	if err := json.NewDecoder(r.Body).Decode(&channel); err != nil {
		log.Printf("Error decoding notification channel creation request: %v", err)
		
		// Audit log failed notification channel creation attempt
		s.auditLogger.LogAction(
			s.getUserIDFromContext(r),
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"create_notification_channel",
			"notification_channels",
			0,
			false,
			"Invalid request body: "+err.Error(),
		)
		
		s.errorHandler.HandleBadRequest(w, r, err, "Invalid request body")
		return
	}
	
	// Validate notification channel data
	// TODO: Add validation for notification channels
	
	if err := s.createNotificationChannel(&channel); err != nil {
		log.Printf("Error creating notification channel: %v", err)
		
		// Audit log failed notification channel creation attempt
		s.auditLogger.LogAction(
			s.getUserIDFromContext(r),
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"create_notification_channel",
			"notification_channels",
			channel.ID,
			false,
			"Database error: "+err.Error(),
		)
		
		s.errorHandler.HandleInternalServerError(w, r, err, "Failed to create notification channel")
		return
	}
	
	log.Printf("Notification channel created successfully: ID=%d, Name=%s", channel.ID, channel.Name)
	
	// Audit log successful notification channel creation
	s.auditLogger.LogAction(
		s.getUserIDFromContext(r),
		s.getIPAddress(r),
		r.Header.Get("User-Agent"),
		"create_notification_channel",
		"notification_channels",
		channel.ID,
		true,
		"Notification channel created successfully",
	)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(channel)
}

// handleGetNotificationChannel returns details for a specific notification channel
func (s *Server) handleGetNotificationChannel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	channelID, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Printf("Invalid notification channel ID provided: %v", err)
		
		// Audit log failed notification channel retrieval attempt
		s.auditLogger.LogAction(
			s.getUserIDFromContext(r),
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"get_notification_channel",
			"notification_channels",
			0,
			false,
			"Invalid notification channel ID: "+err.Error(),
		)
		
		s.errorHandler.HandleBadRequest(w, r, err, "Invalid notification channel ID")
		return
	}
	
	channel, err := s.getNotificationChannelByID(channelID)
	if err != nil {
		log.Printf("Notification channel not found with ID %d: %v", channelID, err)
		
		// Audit log failed notification channel retrieval attempt
		s.auditLogger.LogAction(
			s.getUserIDFromContext(r),
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"get_notification_channel",
			"notification_channels",
			channelID,
			false,
			"Notification channel not found: "+err.Error(),
		)
		
		s.errorHandler.HandleNotFound(w, r, err, "Notification channel not found")
		return
	}
	
	log.Printf("Retrieved notification channel details: ID=%d, Name=%s", channel.ID, channel.Name)
	
	// Audit log successful notification channel retrieval
	s.auditLogger.LogAction(
		s.getUserIDFromContext(r),
		s.getIPAddress(r),
		r.Header.Get("User-Agent"),
		"get_notification_channel",
		"notification_channels",
		channel.ID,
		true,
		"Notification channel details retrieved successfully",
	)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(channel)
}

// handleUpdateNotificationChannel updates a specific notification channel
func (s *Server) handleUpdateNotificationChannel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	channelID, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Printf("Invalid notification channel ID provided: %v", err)
		
		// Audit log failed notification channel update attempt
		s.auditLogger.LogAction(
			s.getUserIDFromContext(r),
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"update_notification_channel",
			"notification_channels",
			0,
			false,
			"Invalid notification channel ID: "+err.Error(),
		)
		
		s.errorHandler.HandleBadRequest(w, r, err, "Invalid notification channel ID")
		return
	}
	
	// Get existing notification channel
	channel, err := s.getNotificationChannelByID(channelID)
	if err != nil {
		log.Printf("Notification channel not found with ID %d: %v", channelID, err)
		
		// Audit log failed notification channel update attempt
		s.auditLogger.LogAction(
			s.getUserIDFromContext(r),
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"update_notification_channel",
			"notification_channels",
			channelID,
			false,
			"Notification channel not found: "+err.Error(),
		)
		
		s.errorHandler.HandleNotFound(w, r, err, "Notification channel not found")
		return
	}
	
	// Parse update data
	var updatedChannel models.NotificationChannel
	if err := json.NewDecoder(r.Body).Decode(&updatedChannel); err != nil {
		log.Printf("Error decoding notification channel update request: %v", err)
		
		// Audit log failed notification channel update attempt
		s.auditLogger.LogAction(
			s.getUserIDFromContext(r),
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"update_notification_channel",
			"notification_channels",
			channelID,
			false,
			"Invalid request body: "+err.Error(),
		)
		
		s.errorHandler.HandleBadRequest(w, r, err, "Invalid request body")
		return
	}
	
	// Update fields
	channel.Name = updatedChannel.Name
	channel.Type = updatedChannel.Type
	channel.Config = updatedChannel.Config
	channel.Enabled = updatedChannel.Enabled
	
	if err := s.updateNotificationChannel(channel); err != nil {
		log.Printf("Error updating notification channel with ID %d: %v", channelID, err)
		
		// Audit log failed notification channel update attempt
		s.auditLogger.LogAction(
			s.getUserIDFromContext(r),
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"update_notification_channel",
			"notification_channels",
			channelID,
			false,
			"Database error: "+err.Error(),
		)
		
		s.errorHandler.HandleInternalServerError(w, r, err, "Failed to update notification channel")
		return
	}
	
	log.Printf("Notification channel updated successfully: ID=%d, Name=%s", channel.ID, channel.Name)
	
	// Audit log successful notification channel update
	s.auditLogger.LogAction(
		s.getUserIDFromContext(r),
		s.getIPAddress(r),
		r.Header.Get("User-Agent"),
		"update_notification_channel",
		"notification_channels",
		channel.ID,
		true,
		"Notification channel updated successfully",
	)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(channel)
}

// handleDeleteNotificationChannel deletes a specific notification channel
func (s *Server) handleDeleteNotificationChannel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	channelID, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Printf("Invalid notification channel ID provided: %v", err)
		
		// Audit log failed notification channel deletion attempt
		s.auditLogger.LogAction(
			s.getUserIDFromContext(r),
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"delete_notification_channel",
			"notification_channels",
			0,
			false,
			"Invalid notification channel ID: "+err.Error(),
		)
		
		s.errorHandler.HandleBadRequest(w, r, err, "Invalid notification channel ID")
		return
	}
	
	// Check if notification channel exists before deleting
	_, err = s.getNotificationChannelByID(channelID)
	if err != nil {
		log.Printf("Notification channel not found with ID %d: %v", channelID, err)
		
		// Audit log failed notification channel deletion attempt
		s.auditLogger.LogAction(
			s.getUserIDFromContext(r),
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"delete_notification_channel",
			"notification_channels",
			channelID,
			false,
			"Notification channel not found: "+err.Error(),
		)
		
		s.errorHandler.HandleNotFound(w, r, err, "Notification channel not found")
		return
	}
	
	if err := s.deleteNotificationChannel(channelID); err != nil {
		log.Printf("Error deleting notification channel with ID %d: %v", channelID, err)
		
		// Audit log failed notification channel deletion attempt
		s.auditLogger.LogAction(
			s.getUserIDFromContext(r),
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"delete_notification_channel",
			"notification_channels",
			channelID,
			false,
			"Database error: "+err.Error(),
		)
		
		s.errorHandler.HandleInternalServerError(w, r, err, "Failed to delete notification channel")
		return
	}
	
	log.Printf("Notification channel deleted successfully: ID=%d", channelID)
	
	// Audit log successful notification channel deletion
	s.auditLogger.LogAction(
		s.getUserIDFromContext(r),
		s.getIPAddress(r),
		r.Header.Get("User-Agent"),
		"delete_notification_channel",
		"notification_channels",
		channelID,
		true,
		"Notification channel deleted successfully",
	)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Notification channel deleted successfully"})
}