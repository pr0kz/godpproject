package handler

import (
	"ai-review-system/pkg/auth"
	jwtpkg "ai-review-system/pkg/jwt"
	"ai-review-system/services/order/internal/service"
	"github.com/gin-gonic/gin"
	"strconv"
)

type Handler struct {
	s *service.SeckillService
	t *jwtpkg.Manager
}

func New(s *service.SeckillService, t *jwtpkg.Manager) *Handler { return &Handler{s: s, t: t} }
func (h *Handler) Router() *gin.Engine {
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	p := r.Group("", auth.Middleware(h.t))
	p.POST("/coupons", h.coupon)
	p.POST("/seckill/:coupon_id", h.seckill)
	return r
}
func (h *Handler) coupon(c *gin.Context) {
	var q struct {
		Title     string  `json:"title"`
		Stock     int     `json:"stock"`
		Discount  float64 `json:"discount"`
		BeginTime int64   `json:"begin_time"`
		EndTime   int64   `json:"end_time"`
	}
	if e := c.ShouldBindJSON(&q); e != nil {
		c.JSON(400, gin.H{"error": e.Error()})
		return
	}
	v, e := h.s.CreateCoupon(q.Title, q.Stock, q.Discount, q.BeginTime, q.EndTime)
	if e != nil {
		c.JSON(400, gin.H{"error": e.Error()})
		return
	}
	c.JSON(201, v)
}
func (h *Handler) seckill(c *gin.Context) {
	id, e := strconv.ParseUint(c.Param("coupon_id"), 10, 64)
	if e != nil {
		c.JSON(400, gin.H{"error": "invalid coupon"})
		return
	}
	uid, _ := c.Get("user_id")
	if e = h.s.Seckill(c, uid.(uint), uint(id)); e != nil {
		c.JSON(400, gin.H{"error": e.Error()})
		return
	}
	c.JSON(202, gin.H{"message": "accepted"})
}
