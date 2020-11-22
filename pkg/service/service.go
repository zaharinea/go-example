package service

import (
	"context"

	"github.com/zaharinea/go-example/pkg/repository"
)

// IUserService interface
type IUserService interface {
	Create(ctx context.Context, user *repository.User) (string, error)
	List(ctx context.Context, limit int64, offset int64) ([]repository.User, error)
	GetByID(ctx context.Context, userID string) (repository.User, error)
	Update(ctx context.Context, userID string, update repository.UpdateUser) error
	UpdateAndReturn(ctx context.Context, userID string, update repository.UpdateUser) (repository.User, error)
	DeleteByID(ctx context.Context, userID string) error
}

// Service struct
type Service struct {
	User IUserService
}

// NewService returns a new Service struct
func NewService(repos *repository.Repository) *Service {
	return &Service{
		User: NewUserService(repos.User),
	}
}
