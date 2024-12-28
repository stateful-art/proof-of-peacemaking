package mongodb

import (
	"context"
	"proofofpeacemaking/internal/core/domain"

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
	return err
}

func (r *userRepository) FindByAddress(ctx context.Context, address string) (*domain.User, error) {
	var user domain.User
	err := r.db.Collection("users").FindOne(ctx, bson.M{"address": address}).Decode(&user)
	return &user, err
}

func (r *userRepository) UpdateNonce(ctx context.Context, id primitive.ObjectID, nonce int) error {
	_, err := r.db.Collection("users").UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"nonce": nonce}},
	)
	return err
}
