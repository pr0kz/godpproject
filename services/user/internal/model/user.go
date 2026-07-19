package model

type User struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	Username     string `gorm:"uniqueIndex;size:191;not null" json:"username"`
	Email        string `gorm:"uniqueIndex;size:191;not null" json:"email"`
	PasswordHash string `gorm:"size:255;not null" json:"-"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`
}

func (User) TableName() string { return "users" }
