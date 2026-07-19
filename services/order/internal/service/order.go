package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"ai-review-system/services/order/internal/model"
	"ai-review-system/services/order/internal/repository"
	"github.com/redis/go-redis/v9"
)

const seckillLuaScript = `local stock = tonumber(redis.call('GET', KEYS[1])); if stock == nil or stock <= 0 then return 0 end; redis.call('DECR', KEYS[1]); return 1`

type SeckillService struct {
	couponRepo *repository.CouponRepository
	orderRepo  *repository.OrderRepository
	rdb        redis.UniversalClient
	producer   func(context.Context, uint, uint) error
}

func NewSeckillService(couponRepo *repository.CouponRepository, orderRepo *repository.OrderRepository, rdb redis.UniversalClient, producer func(context.Context, uint, uint) error) *SeckillService {
	return &SeckillService{couponRepo: couponRepo, orderRepo: orderRepo, rdb: rdb, producer: producer}
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
	result, err := redis.NewScript(seckillLuaScript).Run(ctx, s.rdb, []string{stockKey}).Int()
	if err != nil {
		return fmt.Errorf("redis lua error: %w", err)
	}
	if result == 0 {
		return errors.New("sold out")
	}
	if err := s.producer(ctx, userID, couponID); err != nil {
		s.rdb.Incr(ctx, stockKey)
		return fmt.Errorf("failed to queue order: %w", err)
	}
	return nil
}

type OrderCreatedEvent struct {
	UserID   uint `json:"user_id"`
	CouponID uint `json:"coupon_id"`
}

func ProcessOrderMessage(repo *repository.OrderRepository) func(context.Context, []byte) error {
	return func(ctx context.Context, payload []byte) error {
		var event OrderCreatedEvent
		if err := decodeOrderCreatedEvent(payload, &event); err != nil {
			return err
		}
		return repo.CreateSeckillOrder(ctx, event.UserID, event.CouponID)
	}
}

func decodeOrderCreatedEvent(payload []byte, event *OrderCreatedEvent) error {
	if err := json.Unmarshal(payload, event); err != nil {
		return fmt.Errorf("decode order event: %w", err)
	}
	if event.UserID == 0 || event.CouponID == 0 {
		return errors.New("order event requires user_id and coupon_id")
	}
	return nil
}
