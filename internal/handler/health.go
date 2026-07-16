package handler

import (
	"context"
	"net/http"
	"time"

	"ai-review-system/internal/database"
	"github.com/gin-gonic/gin"
)

func registerHealthRoutes(r *gin.Engine) {
	r.GET("/ping", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "pong"}) })
	r.GET("/livez", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/health", readinessHandler)
	r.GET("/readyz", readinessHandler)
}

func readinessHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	if database.GetDB() == nil || database.GetRedis() == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready"})
		return
	}
	sqlDB, err := database.GetDB().DB()
	if err != nil || sqlDB.PingContext(ctx) != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready", "dependency": "mysql"})
		return
	}
	if err := database.GetRedis().Ping(ctx).Err(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready", "dependency": "redis"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "ai-review-system", "timestamp": time.Now().UTC().Format(time.RFC3339)})
}
