package repository

import (
	"ai-review-system/internal/database"
	"ai-review-system/internal/model"
)

// UserRepository handles user data access
type UserRepository struct{}

// NewUserRepository creates a new user repository
func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

// Create creates a new user in the database
func (r *UserRepository) Create(user *model.User) error {
	return database.GetDB().Create(user).Error
}

// GetByUsername retrieves a user by username
func (r *UserRepository) GetByUsername(username string) (*model.User, error) {
	var user model.User
	err := database.GetDB().Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(email string) (*model.User, error) {
	var user model.User
	err := database.GetDB().Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(id uint) (*model.User, error) {
	var user model.User
	err := database.GetDB().First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
