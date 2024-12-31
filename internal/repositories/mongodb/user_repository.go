package mongodb

import (
	"context"
	"fmt"
	"log"
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
	log.Printf("[USER_REPO] Creating user with address: %s", user.Address)
	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		log.Printf("[USER_REPO] Error creating user: %v", err)
		return err
	}
	log.Printf("[USER_REPO] Created user with ID: %v", result.InsertedID)
	return nil
}

func (r *userRepository) FindByAddress(ctx context.Context, address string) (*domain.User, error) {
	log.Printf("[USER_REPO] Finding user by address: %s", address)
	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{"address": address}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		log.Printf("[USER_REPO] No user found for address: %s", address)
		return nil, nil
	}
	if err != nil {
		log.Printf("[USER_REPO] Error finding user: %v", err)
		return nil, err
	}
	log.Printf("[USER_REPO] Found user with ID: %s", user.ID.Hex())
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
