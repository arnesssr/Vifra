package main

import (
	"log"

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

	log.Println("Database migration completed successfully")
}