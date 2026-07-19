package model

// Review represents a review for a shop
type Review struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	ShopID    uint   `gorm:"not null;index" json:"shop_id"`
	UserID    uint   `gorm:"not null;index" json:"user_id"`
	Content   string `gorm:"size:1024;not null" json:"content"`
	Score     int    `gorm:"not null;default:5" json:"score"`
	LikeCount int    `gorm:"not null;default:0" json:"like_count"`
	Images    string `gorm:"size:1024" json:"images"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

// TableName specifies the table name for Review model
func (Review) TableName() string {
	return "reviews"
}

// Like represents a like record on a review
type Like struct {
	ID        uint  `gorm:"primaryKey" json:"id"`
	ReviewID  uint  `gorm:"not null;uniqueIndex:idx_review_user" json:"review_id"`
	UserID    uint  `gorm:"not null;uniqueIndex:idx_review_user" json:"user_id"`
	CreatedAt int64 `json:"created_at"`
}

// TableName specifies the table name for Like model
func (Like) TableName() string {
	return "likes"
}
