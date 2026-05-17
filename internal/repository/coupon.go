package repository

import (
	"context"
	"fmt"
	"time"

	"ai-review-system/internal/database"
	"ai-review-system/internal/model"
)

const couponStockKeyFmt = "coupon_stock:%d"
const couponStockTTL = 24 * time.Hour

// CouponRepository handles coupon data access
type CouponRepository struct{}

func NewCouponRepository() *CouponRepository {
	return &CouponRepository{}
}

// Create creates a new coupon and initialises its Redis stock
func (r *CouponRepository) Create(coupon *model.Coupon) error {
	if err := database.GetDB().Create(coupon).Error; err != nil {
		return err
	}
	// Pre-load stock into Redis
	ctx := context.Background()
	key := fmt.Sprintf(couponStockKeyFmt, coupon.ID)
	database.GetRedis().Set(ctx, key, coupon.Stock, couponStockTTL)
	return nil
}

// GetByID retrieves a coupon by ID
func (r *CouponRepository) GetByID(id uint) (*model.Coupon, error) {
	var coupon model.Coupon
	if err := database.GetDB().First(&coupon, id).Error; err != nil {
		return nil, err
	}
	return &coupon, nil
}

// DecrStock decrements coupon stock in DB (called after Kafka consumer)
func (r *CouponRepository) DecrStock(id uint) error {
	return database.GetDB().
		Model(&model.Coupon{}).
		Where("id = ? AND stock > 0", id).
		UpdateColumn("stock", database.GetDB().Raw("stock - 1")).
		Error
}

// EnsureRedisStock ensures the Redis stock key exists; loads from DB if missing
func (r *CouponRepository) EnsureRedisStock(ctx context.Context, id uint) error {
	key := fmt.Sprintf(couponStockKeyFmt, id)
	exists, err := database.GetRedis().Exists(ctx, key).Result()
	if err != nil {
		return err
	}
	if exists == 0 {
		coupon, err := r.GetByID(id)
		if err != nil {
			return err
		}
		database.GetRedis().Set(ctx, key, coupon.Stock, couponStockTTL)
	}
	return nil
}
