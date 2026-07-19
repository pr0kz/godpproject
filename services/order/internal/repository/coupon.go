package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"ai-review-system/services/order/internal/model"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const couponStockKeyFmt = "coupon_stock:%d"
const couponStockTTL = 24 * time.Hour

type CouponRepository struct {
	db  *gorm.DB
	rdb redis.UniversalClient
}

func NewCouponRepository(db *gorm.DB, rdb redis.UniversalClient) *CouponRepository {
	return &CouponRepository{db: db, rdb: rdb}
}
func (r *CouponRepository) Create(coupon *model.Coupon) error {
	if err := r.db.Create(coupon).Error; err != nil {
		return err
	}
	return r.rdb.Set(context.Background(), fmt.Sprintf(couponStockKeyFmt, coupon.ID), coupon.Stock, couponStockTTL).Err()
}
func (r *CouponRepository) GetByID(id uint) (*model.Coupon, error) {
	var coupon model.Coupon
	if err := r.db.First(&coupon, id).Error; err != nil {
		return nil, err
	}
	return &coupon, nil
}
func (r *CouponRepository) DecrStock(id uint) error {
	result := r.db.Model(&model.Coupon{}).Where("id = ? AND stock > 0", id).UpdateColumn("stock", gorm.Expr("stock - 1"))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return errors.New("sold out")
	}
	return nil
}
func (r *CouponRepository) EnsureRedisStock(ctx context.Context, id uint) error {
	key := fmt.Sprintf(couponStockKeyFmt, id)
	exists, err := r.rdb.Exists(ctx, key).Result()
	if err != nil {
		return err
	}
	if exists == 0 {
		coupon, err := r.GetByID(id)
		if err != nil {
			return err
		}
		return r.rdb.Set(ctx, key, coupon.Stock, couponStockTTL).Err()
	}
	return nil
}
