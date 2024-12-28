package mongodb

import (
	"context"
	"fmt"
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type userRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) ports.UserRepository {
	return &userRepository{
		collection: db.Collection("users"),
	}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	_, err := r.collection.InsertOne(ctx, user)
	return err
}

func (r *userRepository) FindByAddress(ctx context.Context, address string) (*domain.User, error) {
	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{"address": address}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	filter := bson.M{"_id": user.ID}
	update := bson.M{"$set": user}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *userRepository) UpdateNonce(ctx context.Context, id primitive.ObjectID, nonce int) error {
	result, err := r.collection.UpdateOne(
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
