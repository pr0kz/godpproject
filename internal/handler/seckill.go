package handler

import (
	"net/http"
	"strconv"

	"ai-review-system/internal/service"
	"ai-review-system/pkg/jwt"
	"github.com/gin-gonic/gin"
)

type CreateCouponRequest struct {
	Title     string  `json:"title" binding:"required"`
	Stock     int     `json:"stock" binding:"required,min=1"`
	Discount  float64 `json:"discount" binding:"required"`
	BeginTime int64   `json:"begin_time" binding:"required"`
	EndTime   int64   `json:"end_time" binding:"required"`
}

func CreateCouponHandler(c *gin.Context) {
	var req CreateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	coupon, err := service.NewSeckillService().CreateCoupon(req.Title, req.Stock, req.Discount, req.BeginTime, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "coupon created successfully", "coupon": coupon})
}
func SeckillHandler(c *gin.Context) {
	couponID, err := strconv.ParseUint(c.Param("coupon_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid coupon id"})
		return
	}
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	if err := service.NewSeckillService().Seckill(c.Request.Context(), userID.(uint), uint(couponID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"message": "seckill request accepted"})
}
func registerSeckillRoutes(r *gin.Engine, tokens *jwt.Manager) {
	protected := r.Group("")
	protected.Use(authMiddlewareFunc(tokens))
	protected.POST("/coupons", CreateCouponHandler)
	protected.POST("/seckill/:coupon_id", SeckillHandler)
}
