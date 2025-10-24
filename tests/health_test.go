package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/username/vps-monitor/internal/api"
	"github.com/username/vps-monitor/internal/config"
	"github.com/username/vps-monitor/internal/database"
)

func TestHealthCheck(t *testing.T) {
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
	
	// Create a test request
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a response recorder
	rr := httptest.NewRecorder()
	
	// Call the handler
	server.Router().ServeHTTP(rr, req)
	
	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	
	// Check the response body
	expected := `{"message":"VPS Monitor API is running","status":"healthy"}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}