package handler

import (
	"strings"

	"ai-review-system/internal/middleware"
	jwtpkg "ai-review-system/pkg/jwt"
	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	User   *UserHandler
	Tokens *jwtpkg.Manager
}

func NewRouter(deps Dependencies) *gin.Engine { return NewRouterForService("", deps) }
func NewRouterForService(serviceRole string, deps Dependencies) *gin.Engine {
	r := gin.Default()
	registerHealthRoutes(r)
	switch strings.ToLower(serviceRole) {
	case "user":
		registerUserRoutes(r, deps)
	case "shop":
		registerShopRoutes(r, deps.Tokens)
	case "review":
		registerReviewRoutes(r, deps.Tokens)
	case "order", "seckill":
		registerSeckillRoutes(r, deps.Tokens)
	default:
		registerUserRoutes(r, deps)
		registerShopRoutes(r, deps.Tokens)
		registerReviewRoutes(r, deps.Tokens)
		registerSeckillRoutes(r, deps.Tokens)
	}
	return r
}
func registerUserRoutes(r *gin.Engine, deps Dependencies) {
	r.POST("/register", deps.User.Register)
	r.POST("/login", deps.User.Login)
	protected := r.Group("")
	protected.Use(middleware.AuthMiddleware(deps.Tokens))
	protected.GET("/user/profile", deps.User.Profile)
}
func authMiddlewareFunc(tokens *jwtpkg.Manager) gin.HandlerFunc {
	return middleware.AuthMiddleware(tokens)
}
