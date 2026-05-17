package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"ai-review-system/internal/database"
	"ai-review-system/internal/model"
	"ai-review-system/internal/repository"

	"github.com/redis/go-redis/v9"
)

// Lua script: atomically check stock > 0 and decrement by 1
// Returns 1 on success, 0 if out of stock
const seckillLuaScript = `
local stock = tonumber(redis.call('GET', KEYS[1]))
if stock == nil or stock <= 0 then
    return 0
end
redis.call('DECR', KEYS[1])
return 1
`

// SeckillService handles seckill business logic
type SeckillService struct {
	couponRepo *repository.CouponRepository
	orderRepo  *repository.OrderRepository
}

func NewSeckillService() *SeckillService {
	return &SeckillService{
		couponRepo: repository.NewCouponRepository(),
		orderRepo:  repository.NewOrderRepository(),
	}
}

// CreateCoupon creates a seckill coupon (admin operation)
func (s *SeckillService) CreateCoupon(title string, stock int, discount float64, beginTime, endTime int64) (*model.Coupon, error) {
	if title == "" || stock <= 0 {
		return nil, errors.New("title and stock are required")
	}
	if endTime <= beginTime {
		return nil, errors.New("end_time must be after begin_time")
	}
	now := time.Now().Unix()
	coupon := &model.Coupon{
		Title:     title,
		Stock:     stock,
		MaxStock:  stock,
		Discount:  discount,
		BeginTime: beginTime,
		EndTime:   endTime,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.couponRepo.Create(coupon); err != nil {
		return nil, err
	}
	return coupon, nil
}

// Seckill performs the seckill operation for a user
func (s *SeckillService) Seckill(ctx context.Context, userID, couponID uint) error {
	// 1. Load coupon and validate time window
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

	// 2. Check if user already ordered (one order per user per coupon)
	exists, err := s.orderRepo.ExistsForUser(userID, couponID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("you have already purchased this coupon")
	}

	// 3. Ensure Redis stock key exists
	if err := s.couponRepo.EnsureRedisStock(ctx, couponID); err != nil {
		return err
	}

	// 4. Atomically decrement stock via Lua script
	rdb := database.GetRedis()
	script := redis.NewScript(seckillLuaScript)
	stockKey := fmt.Sprintf("coupon_stock:%d", couponID)

	result, err := script.Run(ctx, rdb, []string{stockKey}).Int()
	if err != nil {
		return fmt.Errorf("redis lua error: %w", err)
	}
	if result == 0 {
		return errors.New("sold out")
	}

	// 5. Send async order creation message to Kafka
	if err := database.SendOrderMessage(ctx, userID, couponID); err != nil {
		// Rollback Redis stock on Kafka failure
		rdb.Incr(ctx, stockKey)
		return fmt.Errorf("failed to queue order: %w", err)
	}

	return nil
}

// ProcessOrderMessage is the Kafka consumer handler: creates order in DB
func ProcessOrderMessage(userID, couponID uint) {
	orderRepo := repository.NewOrderRepository()
	couponRepo := repository.NewCouponRepository()

	// Guard: duplicate check
	exists, err := orderRepo.ExistsForUser(userID, couponID)
	if err != nil || exists {
		log.Printf("[seckill] Order already exists or error: uid=%d cid=%d err=%v", userID, couponID, err)
		return
	}

	// Decrement DB stock
	if err := couponRepo.DecrStock(couponID); err != nil {
		log.Printf("[seckill] Failed to decrement stock: %v", err)
		return
	}

	// Create order
	order, err := orderRepo.Create(userID, couponID, 0)
	if err != nil {
		log.Printf("[seckill] Failed to create order: %v", err)
		return
	}
	log.Printf("[seckill] Order created: id=%d uid=%d cid=%d", order.ID, userID, couponID)
}
