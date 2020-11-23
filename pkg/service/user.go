package service

import (
	"context"

	"github.com/zaharinea/go-example/pkg/repository"
)

// UserService struct
type UserService struct {
	repo repository.IUserRepository
}

// NewUserService returns a new UserService struct
func NewUserService(repo repository.IUserRepository) *UserService {
	return &UserService{repo: repo}
}

//Create method
func (s *UserService) Create(ctx context.Context, user *repository.User) error {
	return s.repo.Create(ctx, user)
}

//List method
func (s *UserService) List(ctx context.Context, limit int64, offset int64) ([]*repository.User, error) {
	return s.repo.List(ctx, limit, offset)
}

//GetByID method
func (s *UserService) GetByID(ctx context.Context, userID string) (*repository.User, error) {
	return s.repo.GetByID(ctx, userID)
}

//Update method
func (s *UserService) Update(ctx context.Context, userID string, update repository.UpdateUser) error {
	return s.repo.Update(ctx, userID, update)
}

//UpdateAndReturn method
func (s *UserService) UpdateAndReturn(ctx context.Context, userID string, update repository.UpdateUser) (*repository.User, error) {
	return s.repo.UpdateAndReturn(ctx, userID, update)
}

//DeleteByID method
func (s *UserService) DeleteByID(ctx context.Context, userID string) error {
	return s.repo.DeleteByID(ctx, userID)
}
