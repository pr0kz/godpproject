package repository

import (
	"ai-review-system/internal/database"
	"ai-review-system/internal/model"
	"time"
)

// OrderRepository handles order data access
type OrderRepository struct{}

func NewOrderRepository() *OrderRepository {
	return &OrderRepository{}
}

// Create creates a new order in DB
func (r *OrderRepository) Create(userID, couponID uint, amount float64) (*model.Order, error) {
	now := time.Now().Unix()
	order := &model.Order{
		UserID:    userID,
		CouponID:  couponID,
		Status:    1,
		Amount:    amount,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := database.GetDB().Create(order).Error; err != nil {
		return nil, err
	}
	return order, nil
}

// ExistsForUser checks if a user already placed an order for a coupon
func (r *OrderRepository) ExistsForUser(userID, couponID uint) (bool, error) {
	var count int64
	err := database.GetDB().Model(&model.Order{}).
		Where("user_id = ? AND coupon_id = ?", userID, couponID).
		Count(&count).Error
	return count > 0, err
}
