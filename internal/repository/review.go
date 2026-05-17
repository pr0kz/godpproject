package repository

import (
	"ai-review-system/internal/database"
	"ai-review-system/internal/model"
)

// ReviewRepository handles review and like data access
type ReviewRepository struct{}

// NewReviewRepository creates a new review repository
func NewReviewRepository() *ReviewRepository {
	return &ReviewRepository{}
}

// Create creates a new review in the database
func (r *ReviewRepository) Create(review *model.Review) error {
	return database.GetDB().Create(review).Error
}

// GetByShopID retrieves reviews for a shop with pagination
func (r *ReviewRepository) GetByShopID(shopID uint, page, pageSize int) ([]model.Review, int64, error) {
	var reviews []model.Review
	var total int64

	db := database.GetDB().Model(&model.Review{}).Where("shop_id = ?", shopID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&reviews).Error; err != nil {
		return nil, 0, err
	}

	return reviews, total, nil
}

// GetByID retrieves a review by ID
func (r *ReviewRepository) GetByID(id uint) (*model.Review, error) {
	var review model.Review
	if err := database.GetDB().First(&review, id).Error; err != nil {
		return nil, err
	}
	return &review, nil
}

// IsLiked checks if a user has liked a review
func (r *ReviewRepository) IsLiked(reviewID, userID uint) (bool, error) {
	var like model.Like
	err := database.GetDB().Where("review_id = ? AND user_id = ?", reviewID, userID).First(&like).Error
	if err != nil {
		if err.Error() == "record not found" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Like adds a like record and increments like_count atomically
func (r *ReviewRepository) Like(reviewID, userID uint, createdAt int64) error {
	tx := database.GetDB().Begin()
	if tx.Error != nil {
		return tx.Error
	}

	like := &model.Like{ReviewID: reviewID, UserID: userID, CreatedAt: createdAt}
	if err := tx.Create(like).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&model.Review{}).Where("id = ?", reviewID).UpdateColumn("like_count", database.GetDB().Raw("like_count + 1")).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// Unlike removes a like record and decrements like_count atomically
func (r *ReviewRepository) Unlike(reviewID, userID uint) error {
	tx := database.GetDB().Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := tx.Where("review_id = ? AND user_id = ?", reviewID, userID).Delete(&model.Like{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&model.Review{}).Where("id = ? AND like_count > 0", reviewID).UpdateColumn("like_count", database.GetDB().Raw("like_count - 1")).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// GetAvgScoreByShopID calculates the average score of reviews for a shop
func (r *ReviewRepository) GetAvgScoreByShopID(shopID uint) (float64, error) {
	var result struct{ Avg float64 }
	err := database.GetDB().Model(&model.Review{}).Select("AVG(score) as avg").Where("shop_id = ?", shopID).Scan(&result).Error
	return result.Avg, err
}
