package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/Tiavina22/lyrify-backend/internal/config"
	"github.com/Tiavina22/lyrify-backend/internal/services"
)

func main() {

	// Load config
	cfg := config.LoadConfig()

	// Connect database
	db := services.NewDatabase(cfg.DatabaseURL)

	_ = db // will be use later when we add handlers and repositories

	r := gin.New()

	// Middlewares
	r.Use(services.LoggerMiddleware())
	r.Use(gin.Recovery())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	log.Printf("Lyrify backend running on :%s", cfg.Port)
	r.Run(":" + cfg.Port)
}
