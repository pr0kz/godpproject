package service

import (
	"errors"
	"time"

	"ai-review-system/internal/model"
	"ai-review-system/internal/repository"
	"ai-review-system/pkg/crypto"
)

// UserService handles user business logic
type UserService struct {
	repo *repository.UserRepository
}

// NewUserService creates a new user service
func NewUserService() *UserService {
	return &UserService{
		repo: repository.NewUserRepository(),
	}
}

// Register creates a new user account
func (s *UserService) Register(username, email, password string) (*model.User, error) {
	// Validate input
	if username == "" || email == "" || password == "" {
		return nil, errors.New("username, email, and password are required")
	}

	// Check if username already exists
	existingUser, _ := s.repo.GetByUsername(username)
	if existingUser != nil {
		return nil, errors.New("username already exists")
	}

	// Check if email already exists
	existingEmail, _ := s.repo.GetByEmail(email)
	if existingEmail != nil {
		return nil, errors.New("email already exists")
	}

	// Hash password
	passwordHash, err := crypto.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Create user
	now := time.Now().Unix()
	user := &model.User{
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates a user and returns the user if credentials are valid
func (s *UserService) Login(username, password string) (*model.User, error) {
	// Get user by username
	user, err := s.repo.GetByUsername(username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	// Verify password
	if !crypto.VerifyPassword(user.PasswordHash, password) {
		return nil, errors.New("invalid username or password")
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(id uint) (*model.User, error) {
	return s.repo.GetByID(id)
}

