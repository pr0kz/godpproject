package handler

import (
	"ai-review-system/pkg/auth"
	jwtpkg "ai-review-system/pkg/jwt"
	"ai-review-system/services/shop/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type Handler struct {
	s *service.ShopService
	t *jwtpkg.Manager
}

func New(s *service.ShopService, t *jwtpkg.Manager) *Handler { return &Handler{s: s, t: t} }
func (h *Handler) Router() *gin.Engine {
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.GET("/shops", h.list)
	r.GET("/shops/:id", h.get)
	r.POST("/shops", auth.Middleware(h.t), h.create)
	return r
}
func (h *Handler) list(c *gin.Context) {
	p, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	z, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	v, n, e := h.s.ListShops(p, z, c.Query("category"))
	if e != nil {
		c.JSON(500, gin.H{"error": e.Error()})
		return
	}
	c.JSON(200, gin.H{"shops": v, "total": n})
}
func (h *Handler) get(c *gin.Context) {
	id, e := strconv.ParseUint(c.Param("id"), 10, 64)
	if e != nil {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}
	v, e := h.s.GetShopByID(uint(id))
	if e != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	c.JSON(200, v)
}
func (h *Handler) create(c *gin.Context) {
	var q struct {
		Name        string  `json:"name"`
		Category    string  `json:"category"`
		Address     string  `json:"address"`
		Description string  `json:"description"`
		Images      string  `json:"images"`
		AvgPrice    float64 `json:"avg_price"`
	}
	if e := c.ShouldBindJSON(&q); e != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": e.Error()})
		return
	}
	v, e := h.s.CreateShop(q.Name, q.Category, q.Address, q.Description, q.Images, q.AvgPrice)
	if e != nil {
		c.JSON(400, gin.H{"error": e.Error()})
		return
	}
	c.JSON(201, v)
}
