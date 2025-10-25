package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/username/vps-monitor/internal/agent"
)

// Config holds agent configuration
type Config struct {
	ServerURL   string
	AgentKey    string
	Interval    time.Duration
	ServerID    int
}

// Metrics represents the metrics collected from the system
type Metrics struct {
	ServerID    int     `json:"server_id"`
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsed  uint64  `json:"memory_used"`
	MemoryTotal uint64  `json:"memory_total"`
	DiskUsed    uint64  `json:"disk_used"`
	DiskTotal   uint64  `json:"disk_total"`
	NetworkIn   uint64  `json:"network_in"`
	NetworkOut  uint64  `json:"network_out"`
	LoadAvg     float64 `json:"load_avg"`
	Timestamp   time.Time `json:"timestamp"`
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	interval := 30 * time.Second
	if intervalStr := os.Getenv("COLLECTION_INTERVAL"); intervalStr != "" {
		if dur, err := time.ParseDuration(intervalStr); err == nil {
			interval = dur
		}
	}

	serverID := 0
	if serverIDStr := os.Getenv("SERVER_ID"); serverIDStr != "" {
		fmt.Sscanf(serverIDStr, "%d", &serverID)
	}

	return &Config{
		ServerURL: os.Getenv("SERVER_URL"),
		AgentKey:  os.Getenv("AGENT_KEY"),
		Interval:  interval,
		ServerID:  serverID,
	}
}

// CollectMetrics collects system metrics
func CollectMetrics(serverID int) (*Metrics, error) {
	// Collect system metrics
	sysMetrics, err := agent.CollectSystemMetrics()
	if err != nil {
		return nil, fmt.Errorf("failed to collect system metrics: %v", err)
	}

	// Convert to API metrics format
	metrics := &Metrics{
		ServerID:    serverID,
		CPUUsage:    sysMetrics.CPUUsage,
		MemoryUsed:  sysMetrics.MemoryUsed,
		MemoryTotal: sysMetrics.MemoryTotal,
		DiskUsed:    sysMetrics.DiskUsed,
		DiskTotal:   sysMetrics.DiskTotal,
		NetworkIn:   0, // TODO: Implement network metrics collection
		NetworkOut:  0, // TODO: Implement network metrics collection
		LoadAvg:     sysMetrics.LoadAvg,
		Timestamp:   time.Now(),
	}

	return metrics, nil
}

// SubmitMetrics submits metrics to the server
func SubmitMetrics(config *Config, metrics *Metrics) error {
	// Convert metrics to JSON
	data, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", config.ServerURL+"/api/v1/metrics", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.AgentKey)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send metrics: %v", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	return nil
}

func main() {
	// Load configuration
	config := LoadConfig()

	// Validate required configuration
	if config.ServerURL == "" {
		log.Fatal("SERVER_URL environment variable is required")
	}
	if config.AgentKey == "" {
		log.Fatal("AGENT_KEY environment variable is required")
	}
	if config.ServerID == 0 {
		log.Fatal("SERVER_ID environment variable is required")
	}

	log.Printf("Starting VPS Monitor Agent")
	log.Printf("Server URL: %s", config.ServerURL)
	log.Printf("Collection interval: %v", config.Interval)

	// Create channel to receive interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start collection loop
	ticker := time.NewTicker(config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Collect metrics
			metrics, err := CollectMetrics(config.ServerID)
			if err != nil {
				log.Printf("Failed to collect metrics: %v", err)
				continue
			}

			// Submit metrics
			if err := SubmitMetrics(config, metrics); err != nil {
				log.Printf("Failed to submit metrics: %v", err)
				continue
			}

			log.Printf("Metrics submitted successfully for server ID %d", config.ServerID)

		case <-sigChan:
			log.Println("Received interrupt signal, shutting down...")
			return
		}
	}
}