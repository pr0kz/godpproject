package model

// Coupon represents a seckill coupon
type Coupon struct {
	ID         uint    `gorm:"primaryKey" json:"id"`
	Title      string  `gorm:"size:128;not null" json:"title"`
	Stock      int     `gorm:"not null;default:0" json:"stock"`
	MaxStock   int     `gorm:"not null;default:0" json:"max_stock"`
	Discount   float64 `gorm:"not null;default:0" json:"discount"`
	BeginTime  int64   `gorm:"not null" json:"begin_time"`
	EndTime    int64   `gorm:"not null" json:"end_time"`
	CreatedAt  int64   `json:"created_at"`
	UpdatedAt  int64   `json:"updated_at"`
}

func (Coupon) TableName() string {
	return "coupons"
}










