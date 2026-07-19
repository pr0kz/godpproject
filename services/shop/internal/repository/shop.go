package repository

import (
	"ai-review-system/services/shop/internal/model"
	"gorm.io/gorm"
)

// ShopRepository owns access to the shop service database.
type ShopRepository struct{ db *gorm.DB }

func NewShopRepository(db *gorm.DB) *ShopRepository     { return &ShopRepository{db: db} }
func (r *ShopRepository) Create(shop *model.Shop) error { return r.db.Create(shop).Error }
func (r *ShopRepository) GetByID(id uint) (*model.Shop, error) {
	var shop model.Shop
	if err := r.db.First(&shop, id).Error; err != nil {
		return nil, err
	}
	return &shop, nil
}
func (r *ShopRepository) List(page, pageSize int, category string) ([]model.Shop, int64, error) {
	var shops []model.Shop
	var total int64
	db := r.db.Model(&model.Shop{})
	if category != "" {
		db = db.Where("category = ?", category)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&shops).Error; err != nil {
		return nil, 0, err
	}
	return shops, total, nil
}
func (r *ShopRepository) UpdateScore(id uint, score float64) error {
	return r.db.Model(&model.Shop{}).Where("id = ?", id).Update("score", score).Error
}
