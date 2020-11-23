package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const usersCollection = "users"

// User struct
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// UpdateUser struct
type UpdateUser struct {
	Name      string    `bson:"name" json:"name"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

// UserRepository struct
type UserRepository struct {
	collection *mongo.Collection
}

// NewUserRepository returns a new UserRepository struct
func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		collection: db.Collection(usersCollection),
	}
}

// Create returns a created user ID
func (r *UserRepository) Create(ctx context.Context, user *User) error {
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	insertResult, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		return err
	}
	user.ID = insertResult.InsertedID.(primitive.ObjectID)
	return nil
}

// List returns User list
func (r *UserRepository) List(ctx context.Context, limit int64, offset int64) ([]*User, error) {
	var results []*User

	opts := options.Find().SetSkip(offset).SetLimit(limit).SetSort(bson.D{bson.E{Key: "_id", Value: 1}})

	cur, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return results, err
	}

	for cur.Next(ctx) {
		var elem User
		err := cur.Decode(&elem)
		if err != nil {
			return results, err
		}
		results = append(results, &elem)
	}

	if err := cur.Err(); err != nil {
		return results, err
	}
	defer cur.Close(ctx)

	return results, nil
}

// GetByID returns a User by ID
func (r *UserRepository) GetByID(ctx context.Context, userID string) (*User, error) {
	var user User

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return &user, mongo.ErrNoDocuments
	}

	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		return &user, err
	}
	return &user, nil
}

// Update returns a updated User
func (r *UserRepository) Update(ctx context.Context, userID string, updateUser UpdateUser) error {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return mongo.ErrNoDocuments
	}
	updateUser.UpdatedAt = time.Now()

	filter := bson.M{"_id": objectID}
	update := bson.D{bson.E{Key: "$set", Value: updateUser}}

	_, err = r.collection.UpdateOne(ctx, filter, update)
	return err
}

// UpdateAndReturn returns a updated User
func (r *UserRepository) UpdateAndReturn(ctx context.Context, userID string, updateUser UpdateUser) (*User, error) {
	var user User

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return &user, mongo.ErrNoDocuments
	}
	updateUser.UpdatedAt = time.Now()

	filter := bson.M{"_id": objectID}
	update := bson.D{bson.E{Key: "$set", Value: updateUser}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After).SetUpsert(false)

	err = r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&user)
	if err != nil {
		return &user, err
	}
	return &user, nil
}

// DeleteByID delete User by ID
func (r *UserRepository) DeleteByID(ctx context.Context, userID string) error {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return mongo.ErrNoDocuments
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}

// DeleteAll delete all
func (r *UserRepository) DeleteAll(ctx context.Context) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{})
	return err
}
