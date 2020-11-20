package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

// IUserRepository interface
type IUserRepository interface {
	Create(ctx context.Context, user *User) (string, error)
	List(ctx context.Context, limit int64, offset int64) ([]*User, error)
	GetByID(ctx context.Context, userID string) (User, error)
	Update(ctx context.Context, userID string, update UpdateUser) error
	UpdateAndReturn(ctx context.Context, userID string, update UpdateUser) (User, error)
	DeleteByID(ctx context.Context, userID string) error
}

// Repository struct
type Repository struct {
	User IUserRepository
}

// NewRepository returns a new Repository struct
func NewRepository(db *mongo.Database) *Repository {
	return &Repository{
		User: NewUserRepository(db),
	}
}
