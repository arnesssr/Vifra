package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"github.com/username/vps-monitor/internal/models"
)

// DB wraps the gorm.DB instance
type DB struct {
	*gorm.DB
}

// Init initializes the database connection
func Init(dsn string) (*DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	
	return &DB{db}, nil
}

// Migrate runs database migrations
func Migrate(db *DB) error {
	// Auto-migrate all models
	err := db.AutoMigrate(
		&models.Server{},
		&models.ServerMetrics{},
		&models.AlertRule{},
		&models.Alert{},
		&models.User{},
	)
	return err
}

// Close closes the database connection
func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}