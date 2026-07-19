package repository

import (
	"context"
	"errors"
	"time"

	"ai-review-system/services/order/internal/model"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type OrderRepository struct{ db *gorm.DB }

func NewOrderRepository(db ...*gorm.DB) *OrderRepository {
	if len(db) > 0 {
		return &OrderRepository{db: db[0]}
	}
	return &OrderRepository{}
}
func (r *OrderRepository) WithDB(db *gorm.DB) *OrderRepository { return &OrderRepository{db: db} }
func (r *OrderRepository) Create(ctx context.Context, userID, couponID uint, amount float64) (*model.Order, error) {
	now := time.Now().Unix()
	order := &model.Order{UserID: userID, CouponID: couponID, Status: 1, Amount: amount, CreatedAt: now, UpdatedAt: now}
	if err := r.db.WithContext(ctx).Create(order).Error; err != nil {
		return nil, err
	}
	return order, nil
}
func (r *OrderRepository) ExistsForUser(ctx context.Context, userID, couponID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Order{}).Where("user_id = ? AND coupon_id = ?", userID, couponID).Count(&count).Error
	return count > 0, err
}
func (r *OrderRepository) CreateSeckillOrder(ctx context.Context, userID, couponID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&model.Order{}).Where("user_id = ? AND coupon_id = ?", userID, couponID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return nil
		}

		result := tx.Model(&model.Coupon{}).Where("id = ? AND stock > 0", couponID).UpdateColumn("stock", gorm.Expr("stock - 1"))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return errors.New("coupon sold out")
		}

		now := time.Now().Unix()
		order := &model.Order{UserID: userID, CouponID: couponID, Status: 1, CreatedAt: now, UpdatedAt: now}
		if err := tx.Create(order).Error; err != nil {
			if isDuplicateKeyError(err) {
				return nil
			}
			return err
		}
		return nil
	})
}

func isDuplicateKeyError(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
