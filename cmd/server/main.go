package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"context"
	"time"

	"github.com/username/vps-monitor/internal/api"
	"github.com/username/vps-monitor/internal/config"
	"github.com/username/vps-monitor/internal/database"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Init(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Run database migrations
	if err := database.Migrate(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Initialize and start server
	server := api.NewServer(cfg, db)
	
	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on %s", cfg.ServerAddress)
		if err := server.Start(); err != nil {
			log.Fatal("Failed to start server:", err)
		}
	}()
	
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	// Graceful shutdown
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := server.Stop(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	
	log.Println("Server exited")
}