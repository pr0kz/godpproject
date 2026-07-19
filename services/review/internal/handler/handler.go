package handler

import (
	"ai-review-system/pkg/auth"
	jwtpkg "ai-review-system/pkg/jwt"
	"ai-review-system/services/review/internal/service"
	"github.com/gin-gonic/gin"
	"strconv"
)

type Handler struct {
	s *service.ReviewService
	t *jwtpkg.Manager
}

func New(s *service.ReviewService, t *jwtpkg.Manager) *Handler { return &Handler{s: s, t: t} }
func (h *Handler) Router() *gin.Engine {
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.GET("/reviews/:shop_id", h.list)
	p := r.Group("/reviews", auth.Middleware(h.t))
	p.POST("", h.create)
	p.POST("/:id/like", h.like)
	return r
}
func (h *Handler) create(c *gin.Context) {
	var q struct {
		ShopID  uint   `json:"shop_id"`
		Content string `json:"content"`
		Score   int    `json:"score"`
		Images  string `json:"images"`
	}
	if e := c.ShouldBindJSON(&q); e != nil {
		c.JSON(400, gin.H{"error": e.Error()})
		return
	}
	uid, _ := c.Get("user_id")
	v, e := h.s.CreateReview(q.ShopID, uid.(uint), q.Content, q.Score, q.Images)
	if e != nil {
		c.JSON(400, gin.H{"error": e.Error()})
		return
	}
	c.JSON(201, v)
}
func (h *Handler) list(c *gin.Context) {
	id, e := strconv.ParseUint(c.Param("shop_id"), 10, 64)
	if e != nil {
		c.JSON(400, gin.H{"error": "invalid shop"})
		return
	}
	p, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	z, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	v, n, e := h.s.GetReviewsByShopID(uint(id), p, z)
	if e != nil {
		c.JSON(500, gin.H{"error": e.Error()})
		return
	}
	c.JSON(200, gin.H{"reviews": v, "total": n})
}
func (h *Handler) like(c *gin.Context) {
	id, e := strconv.ParseUint(c.Param("id"), 10, 64)
	if e != nil {
		c.JSON(400, gin.H{"error": "invalid review"})
		return
	}
	uid, _ := c.Get("user_id")
	v, e := h.s.ToggleLike(uint(id), uid.(uint))
	if e != nil {
		c.JSON(400, gin.H{"error": e.Error()})
		return
	}
	c.JSON(200, gin.H{"liked": v})
}
