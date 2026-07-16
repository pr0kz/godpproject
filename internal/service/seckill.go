package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"ai-review-system/internal/database"
	"ai-review-system/internal/model"
	"ai-review-system/internal/repository"
	"github.com/redis/go-redis/v9"
)

const seckillLuaScript = `local stock = tonumber(redis.call('GET', KEYS[1])); if stock == nil or stock <= 0 then return 0 end; redis.call('DECR', KEYS[1]); return 1`

type SeckillService struct {
	couponRepo *repository.CouponRepository
	orderRepo  *repository.OrderRepository
}

func NewSeckillService() *SeckillService {
	return &SeckillService{couponRepo: repository.NewCouponRepository(), orderRepo: repository.NewOrderRepository(database.GetDB())}
}
func (s *SeckillService) CreateCoupon(title string, stock int, discount float64, beginTime, endTime int64) (*model.Coupon, error) {
	if title == "" || stock <= 0 {
		return nil, errors.New("title and stock are required")
	}
	if endTime <= beginTime {
		return nil, errors.New("end_time must be after begin_time")
	}
	now := time.Now().Unix()
	coupon := &model.Coupon{Title: title, Stock: stock, MaxStock: stock, Discount: discount, BeginTime: beginTime, EndTime: endTime, CreatedAt: now, UpdatedAt: now}
	if err := s.couponRepo.Create(coupon); err != nil {
		return nil, err
	}
	return coupon, nil
}
func (s *SeckillService) Seckill(ctx context.Context, userID, couponID uint) error {
	coupon, err := s.couponRepo.GetByID(couponID)
	if err != nil {
		return errors.New("coupon not found")
	}
	now := time.Now().Unix()
	if now < coupon.BeginTime {
		return errors.New("seckill has not started yet")
	}
	if now > coupon.EndTime {
		return errors.New("seckill has ended")
	}
	exists, err := s.orderRepo.ExistsForUser(ctx, userID, couponID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("you have already purchased this coupon")
	}
	if err := s.couponRepo.EnsureRedisStock(ctx, couponID); err != nil {
		return err
	}
	stockKey := fmt.Sprintf("coupon_stock:%d", couponID)
	result, err := redis.NewScript(seckillLuaScript).Run(ctx, database.GetRedis(), []string{stockKey}).Int()
	if err != nil {
		return fmt.Errorf("redis lua error: %w", err)
	}
	if result == 0 {
		return errors.New("sold out")
	}
	if err := database.SendOrderMessage(ctx, userID, couponID); err != nil {
		database.GetRedis().Incr(ctx, stockKey)
		return fmt.Errorf("failed to queue order: %w", err)
	}
	return nil
}
func ProcessOrderMessage(ctx context.Context, userID, couponID uint) error {
	return repository.NewOrderRepository(database.GetDB()).CreateSeckillOrder(ctx, userID, couponID)
}
