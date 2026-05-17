package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"ai-review-system/internal/service"
)

// CreateCouponRequest represents a coupon creation request
type CreateCouponRequest struct {
	Title     string  `json:"title" binding:"required"`
	Stock     int     `json:"stock" binding:"required,min=1"`
	Discount  float64 `json:"discount"`
	BeginTime int64   `json:"begin_time" binding:"required"`
	EndTime   int64   `json:"end_time" binding:"required"`
}

// CreateCouponHandler handles POST /coupons (admin)
func CreateCouponHandler(c *gin.Context) {
	var req CreateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := service.NewSeckillService()
	coupon, err := svc.CreateCoupon(req.Title, req.Stock, req.Discount, req.BeginTime, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "coupon created successfully",
		"coupon":  coupon,
	})
}

// SeckillHandler handles POST /seckill/:coupon_id
func SeckillHandler(c *gin.Context) {
	idStr := c.Param("coupon_id")
	couponID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid coupon_id"})
		return
	}

	userID, _ := c.Get("user_id")
	ctx := c.Request.Context()

	svc := service.NewSeckillService()
	if err := svc.Seckill(ctx, userID.(uint), uint(couponID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "seckill successful, order is being processed",
	})
}

// registerSeckillRoutes registers seckill-related routes
// 阶段3：秒杀系统
func registerSeckillRoutes(r *gin.Engine) {
	protected := r.Group("")
	protected.Use(authMiddlewareFunc())
	{
		protected.POST("/coupons", CreateCouponHandler)
		protected.POST("/seckill/:coupon_id", SeckillHandler)
	}
}
