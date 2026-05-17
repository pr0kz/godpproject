package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// registerHealthRoutes registers health check routes
func registerHealthRoutes(r *gin.Engine) {
	r.GET("/ping", PingHandler)
	r.GET("/health", HealthHandler)
}

// PingHandler handles GET /ping
func PingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

// HealthHandler handles GET /health
func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"service":   "ai-review-system",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
