package service

import (
	"context"

	"github.com/zaharinea/go-example/pkg/repository"
)

// UserService struct
type UserService struct {
	repo repository.IUserUserRepository
}

// NewUserService returns a new UserService struct
func NewUserService(repo repository.IUserUserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Create(ctx context.Context, user *repository.User) (string, error) {
	return s.repo.Create(ctx, user)
}

func (s *UserService) List(ctx context.Context, limit int64, offset int64) ([]*repository.User, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *UserService) GetByID(ctx context.Context, userID string) (repository.User, error) {
	return s.repo.GetByID(ctx, userID)
}

func (s *UserService) Update(ctx context.Context, userID string, update repository.UpdateUser) error {
	return s.repo.Update(ctx, userID, update)
}

func (s *UserService) DeleteByID(ctx context.Context, userID string) error {
	return s.repo.DeleteByID(ctx, userID)
}
