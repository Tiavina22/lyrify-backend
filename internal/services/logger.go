package services

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)

		log.Printf(
			"%s | %d | %s | %s",
			c.Request.Method,
			c.Writer.Status(),
			duration,
			c.Request.URL.Path,
		)
	}
}
