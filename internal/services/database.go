package services

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/Tiavina22/lyrify-backend/internal/models"
)

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DefaultDatabaseConfig returns default database configuration
func DefaultDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 10 * time.Minute,
	}
}

// NewDatabase creates a new database connection with retry logic and connection pooling
func NewDatabase(dsn string) *gorm.DB {
	return NewDatabaseWithConfig(dsn, DefaultDatabaseConfig())
}

// NewDatabaseWithConfig creates a new database connection with custom configuration
func NewDatabaseWithConfig(dsn string, config *DatabaseConfig) *gorm.DB {
	var db *gorm.DB
	var err error

	// Retry logic with exponential backoff
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})

		if err == nil {
			break
		}

		waitTime := time.Duration(i+1) * 2 * time.Second
		log.Printf("Failed to connect to database (attempt %d/%d): %v. Retrying in %v...",
			i+1, maxRetries, err, waitTime)
		time.Sleep(waitTime)
	}

	if err != nil {
		log.Fatal("Failed to connect to database after retries:", err)
	}

	log.Println("✓ Database connected successfully")

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get database instance:", err)
	}

	configureConnectionPool(sqlDB, config)

	// Run migrations
	autoMigrate(db)

	return db
}

// configureConnectionPool sets up database connection pooling parameters
func configureConnectionPool(sqlDB *sql.DB, config *DatabaseConfig) {
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	log.Printf("✓ Connection pool configured (MaxOpen: %d, MaxIdle: %d)",
		config.MaxOpenConns, config.MaxIdleConns)
}

// autoMigrate runs all database migrations
func autoMigrate(db *gorm.DB) {
	log.Println("Running database migrations...")

	// List of all models to migrate
	models := []interface{}{
		&models.Artist{},
		&models.Song{},
		&models.Lyrics{},
		&models.RequestHistory{},
		&models.PendingRequest{},
		&models.Device{},          // For future notification feature
		&models.NotificationLog{}, // For future notification feature
	}

	err := db.AutoMigrate(models...)
	if err != nil {
		log.Fatal("Migration failed:", err)
	}

	log.Println("✓ Database migrations completed successfully")

	// Add additional indexes and constraints
	addCustomIndexes(db)
}

// addCustomIndexes adds custom indexes for better query performance
func addCustomIndexes(db *gorm.DB) {
	log.Println("Adding custom indexes...")

	// Composite index for song matching (normalized_title + normalized_artist + duration)
	// This index is already defined in the Song model using GORM tags

	// Unique index on lyrics (song_id, version) to prevent duplicate versions
	// This ensures each song can only have one lyrics per version number
	if !db.Migrator().HasIndex(&models.Lyrics{}, "idx_lyrics_song_version") {
		if err := db.Exec(`
			CREATE UNIQUE INDEX IF NOT EXISTS idx_lyrics_song_version 
			ON lyrics(song_id, version)
		`).Error; err != nil {
			log.Printf("Warning: Failed to create unique index on lyrics: %v", err)
		} else {
			log.Println("✓ Added unique index: idx_lyrics_song_version")
		}
	}

	// Unique constraint on pending_requests to avoid duplicates
	// Already defined in the model using uniqueIndex tag

	log.Println("✓ Custom indexes added")
}

// PingDatabase checks if the database connection is alive
func PingDatabase(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// CloseDatabase closes the database connection gracefully
func CloseDatabase(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	log.Println("✓ Database connection closed")
	return nil
}
