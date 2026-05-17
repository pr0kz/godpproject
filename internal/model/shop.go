package model

// Shop represents a shop in the system
type Shop struct {
	ID          uint    `gorm:"primaryKey" json:"id"`
	Name        string  `gorm:"size:128;not null" json:"name"`
	Category    string  `gorm:"size:64;not null" json:"category"`
	Address     string  `gorm:"size:255;not null" json:"address"`
	AvgPrice    float64 `gorm:"not null;default:0" json:"avg_price"`
	Score       float64 `gorm:"not null;default:0" json:"score"`
	Images      string  `gorm:"size:1024" json:"images"`
	Description string  `gorm:"size:512" json:"description"`
	CreatedAt   int64   `json:"created_at"`
	UpdatedAt   int64   `json:"updated_at"`
}

// TableName specifies the table name for Shop model
func (Shop) TableName() string {
	return "shops"
}
