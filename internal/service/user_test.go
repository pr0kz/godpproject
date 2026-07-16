package service

import (
	"context"
	"errors"
	"testing"

	"ai-review-system/internal/model"
	"gorm.io/gorm"
)

type fakeUsers struct {
	byUsername map[string]*model.User
	byEmail    map[string]*model.User
	created    *model.User
	err        error
}

func (f *fakeUsers) Create(_ context.Context, user *model.User) error { f.created = user; return f.err }
func (f *fakeUsers) GetByUsername(_ context.Context, value string) (*model.User, error) {
	if user := f.byUsername[value]; user != nil {
		return user, nil
	}
	return nil, gorm.ErrRecordNotFound
}
func (f *fakeUsers) GetByEmail(_ context.Context, value string) (*model.User, error) {
	if user := f.byEmail[value]; user != nil {
		return user, nil
	}
	return nil, gorm.ErrRecordNotFound
}
func (f *fakeUsers) GetByID(context.Context, uint) (*model.User, error) {
	return nil, errors.New("not found")
}
func TestRegisterHashesPassword(t *testing.T) {
	repo := &fakeUsers{byUsername: map[string]*model.User{}, byEmail: map[string]*model.User{}}
	user, err := NewUserService(repo).Register(context.Background(), "alice", "alice@example.com", "secret12")
	if err != nil {
		t.Fatal(err)
	}
	if user.PasswordHash == "secret12" || repo.created == nil {
		t.Fatal("password was not securely stored")
	}
}
func TestRegisterRejectsDuplicateUsername(t *testing.T) {
	repo := &fakeUsers{byUsername: map[string]*model.User{"alice": {Username: "alice"}}, byEmail: map[string]*model.User{}}
	if _, err := NewUserService(repo).Register(context.Background(), "alice", "new@example.com", "secret12"); err == nil {
		t.Fatal("expected duplicate username error")
	}
}
