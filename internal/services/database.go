package services

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/Tiavina22/lyrify-backend/internal/models"
)

func NewDatabase(dsn string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Database connected")

	autoMigrate(db)

	return db
}

func autoMigrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&models.Song{},
	)

	if err != nil {
		log.Fatal("Migration failed:", err)
	}

	log.Println("Database migrated")
}
