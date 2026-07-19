package service

import (
	"errors"
	"time"

	"ai-review-system/services/review/internal/model"
	"ai-review-system/services/review/internal/repository"
)

// ReviewService handles review business logic.
type ReviewService struct {
	reviewRepo *repository.ReviewRepository
	publish    func(string, *model.Review) error
}

// NewReviewService creates a review service with its own database repository.
func NewReviewService(reviewRepo *repository.ReviewRepository, publish func(string, *model.Review) error) *ReviewService {
	return &ReviewService{reviewRepo: reviewRepo, publish: publish}
}

// CreateReview creates a new review for a shop
func (s *ReviewService) CreateReview(shopID, userID uint, content string, score int, images string) (*model.Review, error) {
	if content == "" {
		return nil, errors.New("content is required")
	}
	if score < 1 || score > 5 {
		return nil, errors.New("score must be between 1 and 5")
	}

	now := time.Now().Unix()
	review := &model.Review{
		ShopID:    shopID,
		UserID:    userID,
		Content:   content,
		Score:     score,
		Images:    images,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.reviewRepo.Create(review); err != nil {
		return nil, err
	}
	if s.publish != nil {
		if err := s.publish("review.created", review); err != nil {
			return nil, err
		}
	}
	return review, nil
}

// GetReviewsByShopID retrieves reviews for a shop with pagination
func (s *ReviewService) GetReviewsByShopID(shopID uint, page, pageSize int) ([]model.Review, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 10
	}
	return s.reviewRepo.GetByShopID(shopID, page, pageSize)
}

// ToggleLike toggles the like status of a review for a user
func (s *ReviewService) ToggleLike(reviewID, userID uint) (bool, error) {
	// Verify review exists
	if _, err := s.reviewRepo.GetByID(reviewID); err != nil {
		return false, errors.New("review not found")
	}

	liked, err := s.reviewRepo.IsLiked(reviewID, userID)
	if err != nil {
		return false, err
	}

	now := time.Now().Unix()
	if liked {
		// Already liked ?unlike
		if err := s.reviewRepo.Unlike(reviewID, userID); err != nil {
			return false, err
		}
		return false, nil
	}

	// Not liked ?like
	if err := s.reviewRepo.Like(reviewID, userID, now); err != nil {
		return false, err
	}
	return true, nil
}
