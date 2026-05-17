package repository

import (
	"ai-review-system/internal/database"
	"ai-review-system/internal/model"
)

// ShopRepository handles shop data access
type ShopRepository struct{}

// NewShopRepository creates a new shop repository
func NewShopRepository() *ShopRepository {
	return &ShopRepository{}
}

// Create creates a new shop in the database
func (r *ShopRepository) Create(shop *model.Shop) error {
	return database.GetDB().Create(shop).Error
}

// GetByID retrieves a shop by ID
func (r *ShopRepository) GetByID(id uint) (*model.Shop, error) {
	var shop model.Shop
	if err := database.GetDB().First(&shop, id).Error; err != nil {
		return nil, err
	}
	return &shop, nil
}

// List retrieves shops with pagination and optional category filter
func (r *ShopRepository) List(page, pageSize int, category string) ([]model.Shop, int64, error) {
	var shops []model.Shop
	var total int64

	db := database.GetDB().Model(&model.Shop{})
	if category != "" {
		db = db.Where("category = ?", category)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&shops).Error; err != nil {
		return nil, 0, err
	}

	return shops, total, nil
}

// UpdateScore updates the average score of a shop
func (r *ShopRepository) UpdateScore(id uint, score float64) error {
	return database.GetDB().Model(&model.Shop{}).Where("id = ?", id).Update("score", score).Error
}


