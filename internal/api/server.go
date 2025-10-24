package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/username/vps-monitor/internal/config"
	"github.com/username/vps-monitor/internal/database"
	"github.com/username/vps-monitor/internal/models"
)

// Server represents the HTTP server
type Server struct {
	config *config.Config
	db     *database.DB
	router *mux.Router
	httpServer *http.Server
}

// NewServer creates a new Server instance
func NewServer(cfg *config.Config, db *database.DB) *Server {
	router := mux.NewRouter()
	
	server := &Server{
		config: cfg,
		db:     db,
		router: router,
	}
	
	server.setupRoutes()
	
	return server
}

// setupRoutes configures the HTTP routes
func (s *Server) setupRoutes() {
	// API v1 routes
	api := s.router.PathPrefix("/api/v1").Subrouter()
	
	// Auth routes
	api.HandleFunc("/auth/login", s.handleLogin).Methods("POST")
	api.HandleFunc("/auth/logout", s.handleLogout).Methods("POST")
	
	// Server routes
	api.HandleFunc("/servers", s.handleGetServers).Methods("GET")
	api.HandleFunc("/servers", s.handleCreateServer).Methods("POST")
	api.HandleFunc("/servers/{id}", s.handleGetServer).Methods("GET")
	api.HandleFunc("/servers/{id}", s.handleUpdateServer).Methods("PUT")
	api.HandleFunc("/servers/{id}", s.handleDeleteServer).Methods("DELETE")
	
	// Metrics routes
	api.HandleFunc("/servers/{id}/metrics", s.handleGetMetrics).Methods("GET")
	api.HandleFunc("/servers/{id}/metrics/history", s.handleGetMetricsHistory).Methods("GET")
	
	// Agent endpoint for metrics submission
	api.HandleFunc("/metrics", s.handlePostMetrics).Methods("POST")
	
	// Alert routes
	api.HandleFunc("/alerts", s.handleGetAlerts).Methods("GET")
	api.HandleFunc("/alerts", s.handleCreateAlert).Methods("POST")
	api.HandleFunc("/alerts/{id}", s.handleGetAlert).Methods("GET")
	api.HandleFunc("/alerts/{id}", s.handleUpdateAlert).Methods("PUT")
	api.HandleFunc("/alerts/{id}", s.handleDeleteAlert).Methods("DELETE")
	
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