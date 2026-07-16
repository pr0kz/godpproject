package handler

import (
	"net/http"

	"ai-review-system/internal/service"
	jwtpkg "ai-review-system/pkg/jwt"
	"github.com/gin-gonic/gin"
)

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserHandler struct {
	users  *service.UserService
	tokens *jwtpkg.Manager
}

func NewUserHandler(users *service.UserService, tokens *jwtpkg.Manager) *UserHandler {
	return &UserHandler{users: users, tokens: tokens}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.users.Register(c.Request.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "user registered successfully", "user": gin.H{"id": user.ID, "username": user.Username, "email": user.Email}})
}
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.users.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	token, err := h.tokens.GenerateToken(user.ID, user.Username, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "login successful", "token": token, "user": gin.H{"id": user.ID, "username": user.Username, "email": user.Email}})
}
func (h *UserHandler) Profile(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	user, err := h.users.GetUserByID(c.Request.Context(), userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": gin.H{"id": user.ID, "username": user.Username, "email": user.Email}})
}
