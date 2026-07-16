package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"ai-review-system/internal/service"
	"ai-review-system/pkg/jwt"
)

// CreateShopRequest represents a shop creation request
type CreateShopRequest struct {
	Name        string  `json:"name" binding:"required"`
	Category    string  `json:"category" binding:"required"`
	Address     string  `json:"address" binding:"required"`
	Description string  `json:"description"`
	Images      string  `json:"images"`
	AvgPrice    float64 `json:"avg_price"`
}

// ListShopsHandler handles GET /shops
func ListShopsHandler(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	category := c.Query("category")

	shopService := service.NewShopService()
	shops, total, err := shopService.ListShops(page, pageSize, category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"shops":     shops,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetShopHandler handles GET /shops/:id
func GetShopHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid shop id"})
		return
	}

	shopService := service.NewShopService()
	shop, err := shopService.GetShopByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "shop not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"shop": shop})
}

// CreateShopHandler handles POST /shops
func CreateShopHandler(c *gin.Context) {
	var req CreateShopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	shopService := service.NewShopService()
	shop, err := shopService.CreateShop(req.Name, req.Category, req.Address, req.Description, req.Images, req.AvgPrice)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "shop created successfully",
		"shop":    shop,
	})
}

// registerShopRoutes registers shop-related routes
// 阶段2：商铺系统
func registerShopRoutes(r *gin.Engine, tokens *jwt.Manager) {
	// Public routes
	r.GET("/shops", ListShopsHandler)
	r.GET("/shops/:id", GetShopHandler)

	// Protected routes (require JWT)
	protected := r.Group("")
	protected.Use(authMiddlewareFunc(tokens))
	{
		protected.POST("/shops", CreateShopHandler)
	}
}
