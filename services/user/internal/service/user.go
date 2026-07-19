package service

import (
	"ai-review-system/pkg/crypto"
	"ai-review-system/services/user/internal/model"
	"ai-review-system/services/user/internal/repository"
	"context"
	"errors"
	"gorm.io/gorm"
	"time"
)

type Service struct{ repo repository.Store }

func New(r repository.Store) *Service { return &Service{repo: r} }
func (s *Service) Register(c context.Context, n, e, p string) (*model.User, error) {
	if n == "" || e == "" || p == "" {
		return nil, errors.New("username, email and password are required")
	}
	if _, x := s.repo.GetByUsername(c, n); x == nil {
		return nil, errors.New("username already exists")
	} else if !errors.Is(x, gorm.ErrRecordNotFound) {
		return nil, x
	}
	if _, x := s.repo.GetByEmail(c, e); x == nil {
		return nil, errors.New("email already exists")
	} else if !errors.Is(x, gorm.ErrRecordNotFound) {
		return nil, x
	}
	h, x := crypto.HashPassword(p)
	if x != nil {
		return nil, x
	}
	now := time.Now().Unix()
	u := &model.User{Username: n, Email: e, PasswordHash: h, CreatedAt: now, UpdatedAt: now}
	return u, s.repo.Create(c, u)
}
func (s *Service) Login(c context.Context, n, p string) (*model.User, error) {
	u, e := s.repo.GetByUsername(c, n)
	if e != nil || u == nil || !crypto.VerifyPassword(u.PasswordHash, p) {
		return nil, errors.New("invalid username or password")
	}
	return u, nil
}
func (s *Service) Get(c context.Context, id uint) (*model.User, error) { return s.repo.GetByID(c, id) }
