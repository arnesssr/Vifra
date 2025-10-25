package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/username/vps-monitor/internal/models"
	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v4"
)

// handleHealthCheck responds with the server health status
func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status":  "healthy",
		"message": "VPS Monitor API is running",
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// handleLogin handles user login
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		// Audit log failed login attempt with invalid request
		s.auditLogger.LogAction(
			0,
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"login",
			"auth",
			0,
			false,
			"Invalid request body",
		)
		
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Find user by username
	var user models.User
	result := s.db.Where("username = ?", credentials.Username).First(&user)
	if result.Error != nil {
		// Audit log failed login attempt with invalid credentials
		s.auditLogger.LogAction(
			0,
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"login",
			"auth",
			0,
			false,
			"Invalid credentials for username: "+credentials.Username,
		)
		
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	
	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password)); err != nil {
		// Audit log failed login attempt with invalid password
		s.auditLogger.LogAction(
			user.ID,
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"login",
			"auth",
			0,
			false,
			"Invalid password for user ID: "+fmt.Sprintf("%d", user.ID),
		)
		
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	
	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
	})
	
	tokenString, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		// Audit log failed login attempt with token generation error
		s.auditLogger.LogAction(
			user.ID,
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"login",
			"auth",
			0,
			false,
			"Failed to generate token for user ID: "+fmt.Sprintf("%d", user.ID),
		)
		
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}
	
	// Audit log successful login
	s.auditLogger.LogAction(
		user.ID,
		s.getIPAddress(r),
		r.Header.Get("User-Agent"),
		"login",
		"auth",
		0,
		true,
		"Successful login for user ID: "+fmt.Sprintf("%d", user.ID),
	)
	
	// Return token
	response := map[string]string{
		"token": tokenString,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// handleLogout handles user logout
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID := s.getUserIDFromContext(r)
	
	// Audit log logout attempt
	s.auditLogger.LogAction(
		userID,
		s.getIPAddress(r),
		r.Header.Get("User-Agent"),
		"logout",
		"auth",
		0,
		true,
		"User logout for user ID: "+fmt.Sprintf("%d", userID),
	)
	
	// For stateless JWT, we simply return success
	// Client should delete the token
	response := map[string]string{
		"message": "Logged out successfully",
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// handleRefresh handles JWT token refresh
func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (requires authentication)
	userID := s.getUserIDFromContext(r)
	if userID == 0 {
		s.errorHandler.HandleUnauthorized(w, r, fmt.Errorf("authentication required"), "Authentication required")
		return
	}
	
	// Get user role from context
	var userRole string
	if user, ok := r.Context().Value("user").(jwt.MapClaims); ok {
		if role, ok := user["role"].(string); ok {
			userRole = role
		}
	}
	
	// Generate new JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"role":    userRole,
	})
	
	tokenString, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		// Audit log failed token refresh attempt
		s.auditLogger.LogAction(
			userID,
			s.getIPAddress(r),
			r.Header.Get("User-Agent"),
			"refresh_token",
			"auth",
			0,
			false,
			"Failed to generate token: "+err.Error(),
		)
		
		s.errorHandler.HandleInternalServerError(w, r, err, "Failed to generate token")
		return
	}
	
	// Audit log successful token refresh
	s.auditLogger.LogAction(
		userID,
		s.getIPAddress(r),
		r.Header.Get("User-Agent"),
		"refresh_token",
		"auth",
		0,
		true,
		"Token refreshed successfully for user ID: "+fmt.Sprintf("%d", userID),
	)
	
	// Return new token
	response := map[string]string{
		"token": tokenString,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}