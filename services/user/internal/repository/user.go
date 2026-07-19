package repository

import (
	"ai-review-system/services/user/internal/model"
	"context"
	"gorm.io/gorm"
)

type Store interface {
	Create(context.Context, *model.User) error
	GetByUsername(context.Context, string) (*model.User, error)
	GetByEmail(context.Context, string) (*model.User, error)
	GetByID(context.Context, uint) (*model.User, error)
}
type Repository struct{ db *gorm.DB }

func New(db *gorm.DB) *Repository { return &Repository{db: db} }
func (r *Repository) Create(c context.Context, u *model.User) error {
	return r.db.WithContext(c).Create(u).Error
}
func (r *Repository) GetByUsername(c context.Context, v string) (*model.User, error) {
	var u model.User
	if e := r.db.WithContext(c).Where("username = ?", v).First(&u).Error; e != nil {
		return nil, e
	}
	return &u, nil
}
func (r *Repository) GetByEmail(c context.Context, v string) (*model.User, error) {
	var u model.User
	if e := r.db.WithContext(c).Where("email = ?", v).First(&u).Error; e != nil {
		return nil, e
	}
	return &u, nil
}
func (r *Repository) GetByID(c context.Context, id uint) (*model.User, error) {
	var u model.User
	if e := r.db.WithContext(c).First(&u, id).Error; e != nil {
		return nil, e
	}
	return &u, nil
}
