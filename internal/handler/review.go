package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"ai-review-system/internal/service"
	"ai-review-system/pkg/jwt"
)

// CreateReviewRequest represents a review creation request
type CreateReviewRequest struct {
	ShopID  uint   `json:"shop_id" binding:"required"`
	Content string `json:"content" binding:"required"`
	Score   int    `json:"score" binding:"required,min=1,max=5"`
	Images  string `json:"images"`
}

// CreateReviewHandler handles POST /reviews
func CreateReviewHandler(c *gin.Context) {
	var req CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")

	reviewService := service.NewReviewService()
	review, err := reviewService.CreateReview(req.ShopID, userID.(uint), req.Content, req.Score, req.Images)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "review created successfully",
		"review":  review,
	})
}

// GetReviewsHandler handles GET /reviews/:shop_id
func GetReviewsHandler(c *gin.Context) {
	shopIDStr := c.Param("shop_id")
	shopID, err := strconv.ParseUint(shopIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid shop_id"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	reviewService := service.NewReviewService()
	reviews, total, err := reviewService.GetReviewsByShopID(uint(shopID), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reviews":   reviews,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// LikeReviewHandler handles POST /reviews/:id/like
func LikeReviewHandler(c *gin.Context) {
	idStr := c.Param("id")
	reviewID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review id"})
		return
	}

	userID, _ := c.Get("user_id")

	reviewService := service.NewReviewService()
	liked, err := reviewService.ToggleLike(uint(reviewID), userID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	message := "liked successfully"
	if !liked {
		message = "unliked successfully"
	}

	c.JSON(http.StatusOK, gin.H{
		"message": message,
		"liked":   liked,
	})
}

// registerReviewRoutes registers review-related routes
// 阶段2：点评系统
func registerReviewRoutes(r *gin.Engine, tokens *jwt.Manager) {
	// Public routes
	r.GET("/reviews/:shop_id", GetReviewsHandler)

	// Protected routes (require JWT)
	protected := r.Group("")
	protected.Use(authMiddlewareFunc(tokens))
	{
		protected.POST("/reviews", CreateReviewHandler)
		protected.POST("/reviews/:id/like", LikeReviewHandler)
	}
}
