package mongodb

import (
	"context"
	"errors"
	"fmt"
	"proofofpeacemaking/internal/core/domain"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	db *mongo.Database
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	// Check if email already exists
	if user.Email != "" {
		exists, err := r.emailExists(ctx, user.Email)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("email already exists")
		}
	}

	// Check if username already exists
	if user.Username != "" {
		exists, err := r.usernameExists(ctx, user.Username)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("username already exists")
		}
	}

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	result, err := r.db.Collection("users").InsertOne(ctx, user)
	if err != nil {
		return err
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	// Check if email already exists for a different user
	if user.Email != "" {
		exists, err := r.emailExistsForOtherUser(ctx, user.Email, user.ID)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("email already exists")
		}
	}

	// Check if username already exists for a different user
	if user.Username != "" {
		exists, err := r.usernameExistsForOtherUser(ctx, user.Username, user.ID)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("username already exists")
		}
	}

	user.UpdatedAt = time.Now()

	filter := bson.M{"_id": user.ID}
	update := bson.M{"$set": user}

	result, err := r.db.Collection("users").UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var user domain.User
	err = r.db.Collection("users").FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetByAddress(ctx context.Context, address string) (*domain.User, error) {
	var user domain.User
	err := r.db.Collection("users").FindOne(ctx, bson.M{"address": address}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.Collection("users").FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	err := r.db.Collection("users").FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) emailExists(ctx context.Context, email string) (bool, error) {
	count, err := r.db.Collection("users").CountDocuments(ctx, bson.M{"email": email})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *UserRepository) emailExistsForOtherUser(ctx context.Context, email string, userID primitive.ObjectID) (bool, error) {
	count, err := r.db.Collection("users").CountDocuments(ctx, bson.M{
		"email": email,
		"_id":   bson.M{"$ne": userID},
	})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *UserRepository) usernameExists(ctx context.Context, username string) (bool, error) {
	count, err := r.db.Collection("users").CountDocuments(ctx, bson.M{"username": username})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *UserRepository) usernameExistsForOtherUser(ctx context.Context, username string, userID primitive.ObjectID) (bool, error) {
	count, err := r.db.Collection("users").CountDocuments(ctx, bson.M{
		"username": username,
		"_id":      bson.M{"$ne": userID},
	})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *UserRepository) ConnectWallet(ctx context.Context, userID primitive.ObjectID, address string) error {
	// Check if wallet is already connected to another user
	exists, err := r.db.Collection("users").CountDocuments(ctx, bson.M{
		"address": address,
		"_id":     bson.M{"$ne": userID},
	})
	if err != nil {
		return err
	}
	if exists > 0 {
		return errors.New("wallet already connected to another account")
	}

	// Update user with wallet address
	filter := bson.M{"_id": userID}
	update := bson.M{
		"$set": bson.M{
			"address":   address,
			"updatedAt": time.Now(),
		},
	}

	result, err := r.db.Collection("users").UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}

func (r *UserRepository) UpdateNonce(ctx context.Context, id primitive.ObjectID, nonce int) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"nonce":     nonce,
			"updatedAt": time.Now(),
		},
	}

	result, err := r.db.Collection("users").UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update nonce: %w", err)
	}

	if result.MatchedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.db.Collection("users").DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}
