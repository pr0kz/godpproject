package service

import (
	"context"
	"errors"
	"time"

	"ai-review-system/internal/model"
	"ai-review-system/internal/repository"
	"ai-review-system/pkg/crypto"
	"gorm.io/gorm"
)

type UserService struct{ repo repository.UserStore }

func NewUserService(repo repository.UserStore) *UserService { return &UserService{repo: repo} }

func (s *UserService) Register(ctx context.Context, username, email, password string) (*model.User, error) {
	if username == "" || email == "" || password == "" {
		return nil, errors.New("username, email, and password are required")
	}
	if user, err := s.repo.GetByUsername(ctx, username); err == nil && user != nil {
		return nil, errors.New("username already exists")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if user, err := s.repo.GetByEmail(ctx, email); err == nil && user != nil {
		return nil, errors.New("email already exists")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	hash, err := crypto.HashPassword(password)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	user := &model.User{Username: username, Email: email, PasswordHash: hash, CreatedAt: now, UpdatedAt: now}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) Login(ctx context.Context, username, password string) (*model.User, error) {
	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil || !crypto.VerifyPassword(user.PasswordHash, password) {
		return nil, errors.New("invalid username or password")
	}
	return user, nil
}
func (s *UserService) GetUserByID(ctx context.Context, id uint) (*model.User, error) {
	return s.repo.GetByID(ctx, id)
}
