package repository

import (
	"ai-review-system/services/review/internal/model"
	"gorm.io/gorm"
)

type ReviewRepository struct{ db *gorm.DB }

func NewReviewRepository(db *gorm.DB) *ReviewRepository       { return &ReviewRepository{db: db} }
func (r *ReviewRepository) Create(review *model.Review) error { return r.db.Create(review).Error }
func (r *ReviewRepository) GetByShopID(shopID uint, page, pageSize int) ([]model.Review, int64, error) {
	var reviews []model.Review
	var total int64
	db := r.db.Model(&model.Review{}).Where("shop_id = ?", shopID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&reviews).Error; err != nil {
		return nil, 0, err
	}
	return reviews, total, nil
}
func (r *ReviewRepository) GetByID(id uint) (*model.Review, error) {
	var review model.Review
	if err := r.db.First(&review, id).Error; err != nil {
		return nil, err
	}
	return &review, nil
}
func (r *ReviewRepository) IsLiked(reviewID, userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.Like{}).Where("review_id = ? AND user_id = ?", reviewID, userID).Count(&count).Error
	return count > 0, err
}
func (r *ReviewRepository) Like(reviewID, userID uint, createdAt int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&model.Like{ReviewID: reviewID, UserID: userID, CreatedAt: createdAt}).Error; err != nil {
			return err
		}
		return tx.Model(&model.Review{}).Where("id = ?", reviewID).UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
	})
}
func (r *ReviewRepository) Unlike(reviewID, userID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("review_id = ? AND user_id = ?", reviewID, userID).Delete(&model.Like{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}
		return tx.Model(&model.Review{}).Where("id = ? AND like_count > 0", reviewID).UpdateColumn("like_count", gorm.Expr("like_count - 1")).Error
	})
}
func (r *ReviewRepository) GetAvgScoreByShopID(shopID uint) (float64, error) {
	var result struct{ Avg float64 }
	err := r.db.Model(&model.Review{}).Select("COALESCE(AVG(score), 0) AS avg").Where("shop_id = ?", shopID).Scan(&result).Error
	return result.Avg, err
}
