package model

// Order represents a seckill order
type Order struct {
	ID        uint    `gorm:"primaryKey" json:"id"`
	UserID    uint    `gorm:"not null;uniqueIndex:idx_user_coupon" json:"user_id"`
	CouponID  uint    `gorm:"not null;uniqueIndex:idx_user_coupon" json:"coupon_id"`
	Status    int     `gorm:"not null;default:1" json:"status"` // 1=pending 2=paid 3=cancelled
	Amount    float64 `gorm:"not null;default:0" json:"amount"`
	CreatedAt int64   `json:"created_at"`
	UpdatedAt int64   `json:"updated_at"`
}

func (Order) TableName() string {
	return "orders"
}
