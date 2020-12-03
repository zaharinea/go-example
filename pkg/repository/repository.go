package repository

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
)

// IUserRepository interface
type IUserRepository interface {
	Create(ctx context.Context, user *User) error
	List(ctx context.Context, limit int64, offset int64) ([]*User, error)
	GetByID(ctx context.Context, userID string) (*User, error)
	Update(ctx context.Context, userID string, update UpdateUser) error
	UpdateAndReturn(ctx context.Context, userID string, update UpdateUser) (*User, error)
	DeleteByID(ctx context.Context, userID string) error
	DeleteAll(ctx context.Context) error
}

// IAccountRepository interface
type IAccountRepository interface {
	CreateOrUpdate(ctx context.Context, account Account, forceUpdate bool) (*Account, error)
	List(ctx context.Context, limit int64, offset int64) ([]*Account, error)
	GetByExternalID(ctx context.Context, accountExternalID string) (*Account, error)
	DeleteByExternalID(ctx context.Context, accountExternalID string) error
	DeleteAll(ctx context.Context) error
}

// Repository struct
type Repository struct {
	User    IUserRepository
	Account IAccountRepository
}

// IsDuplicateKeyErr DuplicateKey error helper
func (r *Repository) IsDuplicateKeyErr(err error) bool {
	var writeExc mongo.WriteException
	var commandExc mongo.CommandError
	if errors.As(err, &writeExc) {
		for _, we := range writeExc.WriteErrors {
			if we.Code == 11000 {
				return true
			}
		}
	} else if errors.As(err, &commandExc) {
		if commandExc.Code == 11000 {
			return true
		}
	}
	return false
}

// NewRepository returns a new Repository struct
func NewRepository(db *mongo.Database) *Repository {
	return &Repository{
		User: NewUserRepository(db), Account: NewAccountRepository(db),
	}
}
