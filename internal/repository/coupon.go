package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"ai-review-system/internal/database"
	"ai-review-system/internal/model"
	"gorm.io/gorm"
)

const couponStockKeyFmt = "coupon_stock:%d"
const couponStockTTL = 24 * time.Hour

type CouponRepository struct{ db *gorm.DB }

func NewCouponRepository() *CouponRepository                  { return &CouponRepository{db: database.GetDB()} }
func NewCouponRepositoryWithDB(db *gorm.DB) *CouponRepository { return &CouponRepository{db: db} }
func (r *CouponRepository) Create(coupon *model.Coupon) error {
	if err := r.db.Create(coupon).Error; err != nil {
		return err
	}
	return database.GetRedis().Set(context.Background(), fmt.Sprintf(couponStockKeyFmt, coupon.ID), coupon.Stock, couponStockTTL).Err()
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
	exists, err := database.GetRedis().Exists(ctx, key).Result()
	if err != nil {
		return err
	}
	if exists == 0 {
		coupon, err := r.GetByID(id)
		if err != nil {
			return err
		}
		return database.GetRedis().Set(ctx, key, coupon.Stock, couponStockTTL).Err()
	}
	return nil
}
