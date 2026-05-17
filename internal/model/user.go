package model

// User represents a user in the system
type User struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	Username     string `gorm:"uniqueIndex;size:191;not null" json:"username"`
	Email        string `gorm:"uniqueIndex;size:191;not null" json:"email"`
	PasswordHash string `gorm:"size:255;not null" json:"-"` // Never expose password hash
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}
