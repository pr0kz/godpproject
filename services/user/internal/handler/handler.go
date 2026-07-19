package handler

import (
	"ai-review-system/pkg/auth"
	jwtpkg "ai-review-system/pkg/jwt"
	"ai-review-system/services/user/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Handler struct {
	s *service.Service
	t *jwtpkg.Manager
}

func New(s *service.Service, t *jwtpkg.Manager) *Handler { return &Handler{s: s, t: t} }
func (h *Handler) Router() *gin.Engine {
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.POST("/register", h.register)
	r.POST("/login", h.login)
	p := r.Group("/user", auth.Middleware(h.t))
	p.GET("/profile", h.profile)
	return r
}
func (h *Handler) register(c *gin.Context) {
	var q struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}
	if e := c.ShouldBindJSON(&q); e != nil {
		c.JSON(400, gin.H{"error": e.Error()})
		return
	}
	u, e := h.s.Register(c, q.Username, q.Email, q.Password)
	if e != nil {
		c.JSON(400, gin.H{"error": e.Error()})
		return
	}
	c.JSON(201, u)
}
func (h *Handler) login(c *gin.Context) {
	var q struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if e := c.ShouldBindJSON(&q); e != nil {
		c.JSON(400, gin.H{"error": e.Error()})
		return
	}
	u, e := h.s.Login(c, q.Username, q.Password)
	if e != nil {
		c.JSON(401, gin.H{"error": e.Error()})
		return
	}
	t, e := h.t.GenerateToken(u.ID, u.Username, u.Email)
	if e != nil {
		c.JSON(500, gin.H{"error": "token generation failed"})
		return
	}
	c.JSON(200, gin.H{"token": t, "user": u})
}
func (h *Handler) profile(c *gin.Context) {
	id, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	u, e := h.s.Get(c, id.(uint))
	if e != nil {
		c.JSON(404, gin.H{"error": "user not found"})
		return
	}
	c.JSON(200, u)
}
