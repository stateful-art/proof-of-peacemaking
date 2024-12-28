package mongodb

import (
	"context"
	"fmt"
	"proofofpeacemaking/internal/core/domain"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type userRepository struct {
	db *mongo.Database
}

func NewUserRepository(db *mongo.Database) *userRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	_, err := r.db.Collection("users").InsertOne(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *userRepository) FindByAddress(ctx context.Context, address string) (*domain.User, error) {
	var user domain.User
	err := r.db.Collection("users").FindOne(ctx, bson.M{"address": address}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("user not found with address %s", address)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return &user, nil
}

func (r *userRepository) UpdateNonce(ctx context.Context, id primitive.ObjectID, nonce int) error {
	result, err := r.db.Collection("users").UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"nonce":     nonce,
			"updatedAt": primitive.DateTime(time.Now().UnixNano() / int64(time.Millisecond)),
		}},
	)
	if err != nil {
		return fmt.Errorf("failed to update nonce: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found with id %s", id.Hex())
	}
	return nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	result, err := r.db.Collection("users").UpdateOne(
		ctx,
		bson.M{"_id": user.ID},
		bson.M{"$set": bson.M{
			"email":     user.Email,
			"updatedAt": primitive.DateTime(time.Now().UnixNano() / int64(time.Millisecond)),
		}},
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found with id %s", user.ID.Hex())
	}
	return nil
}
