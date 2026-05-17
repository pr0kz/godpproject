package handler

import (
	"strings"

	"github.com/gin-gonic/gin"

	"ai-review-system/internal/middleware"
)

// NewRouter sets up and returns the Gin router with all routes registered.
func NewRouter() *gin.Engine {
	return NewRouterForService("")
}

// NewRouterForService sets up routes for a specific service role.
// Empty serviceRole keeps the original monolith behavior.
func NewRouterForService(serviceRole string) *gin.Engine {
	r := gin.Default()

	registerHealthRoutes(r)

	switch strings.ToLower(serviceRole) {
	case "user":
		registerUserRoutes(r)
	case "shop":
		registerShopRoutes(r)
	case "review":
		registerReviewRoutes(r)
	case "order", "seckill":
		registerSeckillRoutes(r)
	default:
		registerUserRoutes(r)
		registerShopRoutes(r)
		registerReviewRoutes(r)
		registerSeckillRoutes(r)
	}

	return r
}

// authMiddlewareFunc returns the JWT auth middleware handler.
func authMiddlewareFunc() gin.HandlerFunc {
	return middleware.AuthMiddleware()
}
