package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"ai-review-system/services/shop/internal/model"
	"ai-review-system/services/shop/internal/repository"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

const shopCacheTTL = 30 * time.Minute
const shopNullCacheTTL = 2 * time.Minute // empty-value cache to prevent penetration
const shopNullValue = "__null__"

var shopSFGroup singleflight.Group

// ShopService handles shop business logic
type ShopService struct {
	repo *repository.ShopRepository
	rdb  redis.UniversalClient
}

// NewShopService creates a shop service with explicit dependencies.
func NewShopService(repo *repository.ShopRepository, rdb redis.UniversalClient) *ShopService {
	return &ShopService{repo: repo, rdb: rdb}
}

// CreateShop creates a new shop
func (s *ShopService) CreateShop(name, category, address, description, images string, avgPrice float64) (*model.Shop, error) {
	if name == "" || category == "" || address == "" {
		return nil, errors.New("name, category and address are required")
	}

	now := time.Now().Unix()
	shop := &model.Shop{
		Name:        name,
		Category:    category,
		Address:     address,
		Description: description,
		Images:      images,
		AvgPrice:    avgPrice,
		Score:       0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(shop); err != nil {
		return nil, err
	}
	return shop, nil
}

// GetShopByID retrieves a shop by ID.
// Anti-penetration: caches null values for missing shops.
// Anti-breakdown: uses singleflight to collapse concurrent cache misses.
func (s *ShopService) GetShopByID(id uint) (*model.Shop, error) {
	ctx := context.Background()
	rdb := s.rdb
	cacheKey := fmt.Sprintf("shop:%d", id)

	// 1. Try cache first
	if rdb != nil {
		val, err := rdb.Get(ctx, cacheKey).Result()
		if err == nil {
			// Hit: check for null sentinel (cache penetration guard)
			if val == shopNullValue {
				return nil, errors.New("shop not found")
			}
			var shop model.Shop
			if err := json.Unmarshal([]byte(val), &shop); err == nil {
				return &shop, nil
			}
		} else if err != redis.Nil {
			log.Printf("[cache] Redis GET error: %v", err)
		}
	}

	// 2. Cache miss: use singleflight to prevent cache breakdown
	sfKey := fmt.Sprintf("shop_sf:%d", id)
	v, err, _ := shopSFGroup.Do(sfKey, func() (interface{}, error) {
		shop, dbErr := s.repo.GetByID(id)
		if dbErr != nil {
			// Store null sentinel to prevent repeated DB hits (cache penetration)
			if rdb != nil {
				rdb.Set(ctx, cacheKey, shopNullValue, shopNullCacheTTL)
			}
			return nil, errors.New("shop not found")
		}

		// Write back to cache
		if rdb != nil {
			if data, merr := json.Marshal(shop); merr == nil {
				if serr := rdb.Set(ctx, cacheKey, data, shopCacheTTL).Err(); serr != nil {
					log.Printf("[cache] Failed to set shop cache: %v", serr)
				}
			}
		}
		return shop, nil
	})

	if err != nil {
		return nil, err
	}
	return v.(*model.Shop), nil
}

// ListShops retrieves shops with pagination
func (s *ShopService) ListShops(page, pageSize int, category string) ([]model.Shop, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 10
	}
	return s.repo.List(page, pageSize, category)
}

// InvalidateShopCache removes a shop's cache entry
func (s *ShopService) InvalidateShopCache(id uint) {
	ctx := context.Background()
	if s.rdb != nil {
		cacheKey := fmt.Sprintf("shop:%d", id)
		s.rdb.Del(ctx, cacheKey)
	}
}

// UpdateShopScore recalculates and updates a shop's average score
func (s *ShopService) UpdateShopScore(shopID uint, avgScore float64) error {
	if err := s.repo.UpdateScore(shopID, avgScore); err != nil {
		return err
	}
	s.InvalidateShopCache(shopID)
	return nil
}
