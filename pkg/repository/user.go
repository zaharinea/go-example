package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const usersCollection = "users"

// User struct
type User struct {
	ID   primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name string             `bson:"name" json:"name"`
}

// UpdateUser struct
type UpdateUser struct {
	Name string `bson:"name" json:"name"`
}

// UserRepository struct
type UserRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
}

// NewUserRepository returns a new UserRepository struct
func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		db:         db,
		collection: db.Collection(usersCollection),
	}
}

// Create returns a created user ID
func (r UserRepository) Create(ctx context.Context, user *User) (string, error) {
	insertResult, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		return "", err
	}
	user.ID = insertResult.InsertedID.(primitive.ObjectID)
	return insertResult.InsertedID.(primitive.ObjectID).Hex(), nil
}

// List returns User list
func (r UserRepository) List(ctx context.Context, limit int64, offset int64) ([]*User, error) {
	var results []*User

	findOptions := options.Find()
	findOptions.SetSkip(offset).SetLimit(limit).SetSort(bson.D{{"_id", 1}})

	cur, err := r.collection.Find(ctx, bson.D{{}}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var elem User
		err := cur.Decode(&elem)
		if err != nil {
			return nil, err
		}

		results = append(results, &elem)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// GetByID returns a User by ID
func (r UserRepository) GetByID(ctx context.Context, userID string) (User, error) {
	var user User
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return user, err
	}
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		return user, err
	}
	return user, nil
}

// Update returns a updated User
func (r UserRepository) Update(ctx context.Context, userID string, update UpdateUser) error {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.D{{"$set", update}})
	return err
}

// DeleteByID delete User by ID
func (r UserRepository) DeleteByID(ctx context.Context, userID string) error {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}
