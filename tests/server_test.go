package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/username/vps-monitor/internal/api"
	"github.com/username/vps-monitor/internal/config"
	"github.com/username/vps-monitor/internal/database"
	"github.com/username/vps-monitor/internal/models"
)

func TestServerCRUD(t *testing.T) {
	// Create a test database (in memory if possible)
	cfg := &config.Config{
		ServerAddress: ":8080",
		DatabaseURL:   "postgres://user:password@localhost:5432/vpsmonitor?sslmode=disable",
		JWTSecret:     "test-secret",
	}
	
	// Initialize database
	db, err := database.Init(cfg.DatabaseURL)
	if err != nil {
		t.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Create server
	server := api.NewServer(cfg, db)
	
	// Test creating a server
	serverData := models.Server{
		Name:      "Test Server",
		IPAddress: "192.168.1.1",
		Hostname:  "test-server",
		OS:        "Ubuntu 20.04",
	}
	
	jsonData, _ := json.Marshal(serverData)
	req, err := http.NewRequest("POST", "/api/v1/servers", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatal(err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	rr := httptest.NewRecorder()
	server.Router().ServeHTTP(rr, req)
	
	// Check the status code
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}
	
	// Test getting servers
	req, err = http.NewRequest("GET", "/api/v1/servers", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	rr = httptest.NewRecorder()
	server.Router().ServeHTTP(rr, req)
	
	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}