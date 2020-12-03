package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const accountsCollection = "accounts"

// Account struct
type Account struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ExternalID string             `bson:"external_id" json:"external_id"`
	Name       string             `bson:"name" json:"name"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}

// AccountRepository struct
type AccountRepository struct {
	collection *mongo.Collection
}

// NewAccountRepository returns a new AccountRepository struct
func NewAccountRepository(db *mongo.Database) *AccountRepository {
	return &AccountRepository{
		collection: db.Collection(accountsCollection),
	}
}

// CreateOrUpdate returns a updated or created Account
func (r *AccountRepository) CreateOrUpdate(ctx context.Context, account Account, forceUpdate bool) (*Account, error) {
	filter := bson.M{"external_id": account.ExternalID}
	if !forceUpdate {
		filter["updated_at"] = bson.M{"$lt": account.UpdatedAt}
	}

	update := bson.D{bson.E{Key: "$set", Value: account}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After).SetUpsert(true)

	var updatedAccount Account
	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedAccount)
	if err != nil {
		return &updatedAccount, err
	}
	return &updatedAccount, nil
}

// List returns Account list
func (r *AccountRepository) List(ctx context.Context, limit int64, offset int64) ([]*Account, error) {
	var results []*Account

	opts := options.Find().SetSkip(offset).SetLimit(limit).SetSort(bson.D{bson.E{Key: "_id", Value: 1}})

	cur, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return results, err
	}
	err = cur.All(ctx, &results)
	if err != nil {
		return results, err
	}
	return results, nil
}

// GetByExternalID returns a Account by External ID
func (r *AccountRepository) GetByExternalID(ctx context.Context, accountExternalID string) (*Account, error) {
	var Account Account

	err := r.collection.FindOne(ctx, bson.M{"external_id": accountExternalID}).Decode(&Account)
	if err != nil {
		return &Account, err
	}
	return &Account, nil
}

// DeleteByExternalID delete Account by External ID
func (r *AccountRepository) DeleteByExternalID(ctx context.Context, accountExternalID string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"external_id": accountExternalID})
	return err
}

// DeleteAll delete all
func (r *AccountRepository) DeleteAll(ctx context.Context) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{})
	return err
}
