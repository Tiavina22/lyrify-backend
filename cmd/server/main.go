package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Tiavina22/lyrify-backend/internal/config"
	"github.com/Tiavina22/lyrify-backend/internal/features/song"
	"github.com/Tiavina22/lyrify-backend/internal/services"
)

func main() {

	// Load config
	cfg := config.LoadConfig()

	// Connect database
	db := services.NewDatabase(cfg.DatabaseURL)

	// Initialize Gin
	r := gin.New()

	// Middlewares
	r.Use(services.LoggerMiddleware())
	r.Use(gin.Recovery())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// Register API routes
	api := r.Group("/api/v1")
	{
		// Song feature routes
		songHandler := song.NewHandler(db)
		songHandler.RegisterRoutes(api)
	}

	log.Printf("✓ Server started successfully")
	log.Printf("✓ Listening on port %s", cfg.Port)
	log.Printf("✓ Health check: http://localhost:%s/health", cfg.Port)
	log.Printf("✓ API endpoints: http://localhost:%s/api/v1", cfg.Port)

	r.Run(":" + cfg.Port)
}
