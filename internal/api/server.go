package api

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/golang-jwt/jwt/v4"
	"github.com/username/vps-monitor/internal/audit"
	"github.com/username/vps-monitor/internal/config"
	"github.com/username/vps-monitor/internal/crypto"
	"github.com/username/vps-monitor/internal/database"
	"github.com/username/vps-monitor/internal/errors"
	"github.com/username/vps-monitor/internal/models"
	"github.com/username/vps-monitor/internal/ratelimit"
	"github.com/username/vps-monitor/internal/security"
)

// Server represents the HTTP server
type Server struct {
	config *config.Config
	db     *database.DB
	router *mux.Router
	httpServer *http.Server
	rateLimiter *ratelimit.RateLimiter
	auditLogger *audit.Logger
	encryptor *crypto.Encryptor
	errorHandler *errors.ErrorHandler
}

// NewServer creates a new Server instance
func NewServer(cfg *config.Config, db *database.DB) *Server {
	router := mux.NewRouter()
	
	// Initialize rate limiter (100 requests per minute per IP)
	rateLimiter := ratelimit.NewRateLimiter(100, time.Minute)
	
	// Initialize audit logger
	auditLogger := audit.NewLogger()
	
	// Initialize encryptor if encryption key is provided
	var encryptor *crypto.Encryptor
	if encryptionKey, err := cfg.GetEncryptionKey(); err == nil && encryptionKey != nil {
		if enc, err := crypto.NewEncryptor(encryptionKey); err == nil {
			encryptor = enc
		} else {
			log.Printf("Warning: Failed to initialize encryptor: %v", err)
		}
	} else if err != nil {
		log.Printf("Warning: Failed to get encryption key: %v", err)
	}
	
	// Initialize error handler (set debug to false in production)
	errorHandler := errors.NewErrorHandler(false)
	
	server := &Server{
		config: cfg,
		db:     db,
		router: router,
		rateLimiter: rateLimiter,
		auditLogger: auditLogger,
		encryptor: encryptor,
		errorHandler: errorHandler,
	}
	
	server.setupRoutes()
	
	return server
}

// setupRoutes configures the HTTP routes
func (s *Server) setupRoutes() {
	// Apply security headers middleware to all routes
	s.router.Use(security.SecurityHeadersMiddleware)
	
	// Apply CORS middleware to all routes
	s.router.Use(security.CORSMiddleware)
	
	// Apply rate limiting to all routes
	s.router.Use(s.rateLimiter.Middleware)
	
	// API v1 routes
	api := s.router.PathPrefix("/api/v1").Subrouter()
	
	// Public routes
	api.HandleFunc("/auth/login", s.handleLogin).Methods("POST")
	api.HandleFunc("/auth/logout", s.handleLogout).Methods("POST")
	
	// Protected routes - require authentication
	protected := api.PathPrefix("").Subrouter()
	protected.Use(s.JWTMiddleware)
	
	// Server routes
	protected.HandleFunc("/servers", s.handleGetServers).Methods("GET")
	protected.HandleFunc("/servers", s.handleCreateServer).Methods("POST")
	protected.HandleFunc("/servers/{id}", s.handleGetServer).Methods("GET")
	protected.HandleFunc("/servers/{id}", s.handleUpdateServer).Methods("PUT")
	protected.HandleFunc("/servers/{id}", s.handleDeleteServer).Methods("DELETE")
	
	// Metrics routes
	protected.HandleFunc("/servers/{id}/metrics", s.handleGetMetrics).Methods("GET")
	protected.HandleFunc("/servers/{id}/metrics/history", s.handleGetMetricsHistory).Methods("GET")
	
	// Agent endpoint for metrics submission (might be public or have different auth)
	api.HandleFunc("/metrics", s.handlePostMetrics).Methods("POST")
	
	// Alert routes
	protected.HandleFunc("/alerts", s.handleGetAlerts).Methods("GET")
	protected.HandleFunc("/alerts", s.handleCreateAlert).Methods("POST")
	protected.HandleFunc("/alerts/{id}", s.handleGetAlert).Methods("GET")
	protected.HandleFunc("/alerts/{id}", s.handleUpdateAlert).Methods("PUT")
	protected.HandleFunc("/alerts/{id}", s.handleDeleteAlert).Methods("DELETE")
	
	// Health check
	s.router.HandleFunc("/health", s.handleHealthCheck).Methods("GET")
}

// Start begins listening for HTTP requests
func (s *Server) Start() error {
	s.httpServer = &http.Server{
		Addr:         s.config.ServerAddress,
		Handler:      s.router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	
	return s.httpServer.ListenAndServe()
}

// Stop gracefully shuts down the server
func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// getUserIDFromContext extracts user ID from the request context
func (s *Server) getUserIDFromContext(r *http.Request) int {
	if user, ok := r.Context().Value("user").(jwt.MapClaims); ok {
		if userID, ok := user["user_id"].(float64); ok {
			return int(userID)
		}
	}
	return 0
}

// getIPAddress extracts IP address from the request
func (s *Server) getIPAddress(r *http.Request) string {
	// Check X-Forwarded-For header first
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Take the first IP if there are multiple
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

// decryptAgentKey decrypts an agent key if encryption is enabled
func (s *Server) decryptAgentKey(encryptedKey string) (string, error) {
	if s.encryptor == nil {
		return encryptedKey, nil // Not encrypted
	}
	
	decrypted, err := s.encryptor.Decrypt(encryptedKey)
	if err != nil {
		return "", err
	}
	
	return string(decrypted), nil
}

// checkServerAccess verifies that the user has access to the specified server
// For now, we'll implement a simple check that the server exists
// In a more complex system, you might check user permissions or ownership
func (s *Server) checkServerAccess(serverID int) error {
	// Check if server exists
	_, err := s.getServerByID(serverID)
	if err != nil {
		// Return a generic error to prevent enumeration
		return fmt.Errorf("access denied")
	}
	
	return nil
}

// JWTMiddleware validates JWT tokens
func (s *Server) JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Printf("Authentication failed: Authorization header missing for %s %s", r.Method, r.URL.Path)
			s.errorHandler.HandleUnauthorized(w, r, fmt.Errorf("missing authorization header"), "Authentication required")
			return
		}

		// Check if header has Bearer prefix
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			log.Printf("Authentication failed: Invalid authorization header format for %s %s", r.Method, r.URL.Path)
			s.errorHandler.HandleUnauthorized(w, r, fmt.Errorf("invalid authorization header format"), "Authentication required")
			return
		}

		// Extract token
		tokenString := authHeader[7:]

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(s.config.JWTSecret), nil
		})

		if err != nil {
			log.Printf("Authentication failed: Error parsing token for %s %s: %v", r.Method, r.URL.Path, err)
			s.errorHandler.HandleUnauthorized(w, r, err, "Invalid token")
			return
		}

		if !token.Valid {
			log.Printf("Authentication failed: Invalid token for %s %s", r.Method, r.URL.Path)
			s.errorHandler.HandleUnauthorized(w, r, fmt.Errorf("invalid token"), "Invalid token")
			return
		}

		// Add user claims to request context
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			ctx := context.WithValue(r.Context(), "user", claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			log.Printf("Authentication failed: Invalid token claims for %s %s", r.Method, r.URL.Path)
			s.errorHandler.HandleUnauthorized(w, r, fmt.Errorf("invalid token claims"), "Invalid token")
			return
		}
	})
}

// Helper functions for database operations
func (s *Server) getServers() ([]models.Server, error) {
	var servers []models.Server
	result := s.db.Find(&servers)
	return servers, result.Error
}

func (s *Server) getServerByID(id int) (*models.Server, error) {
	var server models.Server
	result := s.db.First(&server, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &server, nil
}

func (s *Server) createServer(server *models.Server) error {
	result := s.db.Create(server)
	return result.Error
}

func (s *Server) updateServer(server *models.Server) error {
	result := s.db.Save(server)
	return result.Error
}

func (s *Server) deleteServer(id int) error {
	result := s.db.Delete(&models.Server{}, id)
	return result.Error
}

// Helper functions for metrics operations
func (s *Server) getMetrics(serverID int) (*models.ServerMetrics, error) {
	var metrics models.ServerMetrics
	result := s.db.Where("server_id = ?", serverID).Order("timestamp desc").First(&metrics)
	if result.Error != nil {
		return nil, result.Error
	}
	return &metrics, nil
}

func (s *Server) getMetricsHistory(serverID int) ([]models.ServerMetrics, error) {
	var metrics []models.ServerMetrics
	result := s.db.Where("server_id = ?", serverID).Order("timestamp desc").Limit(100).Find(&metrics)
	return metrics, result.Error
}

func (s *Server) createMetrics(metrics *models.ServerMetrics) error {
	result := s.db.Create(metrics)
	return result.Error
}

// Helper functions for alert operations
func (s *Server) getAlerts() ([]models.Alert, error) {
	var alerts []models.Alert
	result := s.db.Find(&alerts)
	return alerts, result.Error
}

func (s *Server) getAlertByID(id int) (*models.Alert, error) {
	var alert models.Alert
	result := s.db.First(&alert, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &alert, nil
}

func (s *Server) createAlert(alert *models.Alert) error {
	result := s.db.Create(alert)
	return result.Error
}

func (s *Server) updateAlert(alert *models.Alert) error {
	result := s.db.Save(alert)
	return result.Error
}

func (s *Server) deleteAlert(id int) error {
	result := s.db.Delete(&models.Alert{}, id)
	return result.Error
}

// Helper functions for alert rule operations
func (s *Server) getAlertRuleByID(id int) (*models.AlertRule, error) {
	var alertRule models.AlertRule
	result := s.db.First(&alertRule, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &alertRule, nil
}