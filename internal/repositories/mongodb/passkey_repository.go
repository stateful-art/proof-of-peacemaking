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

type passkeyRepository struct {
	db               *mongo.Database
	credentialsColl  *mongo.Collection
	userPasskeysColl *mongo.Collection
}

// NewPasskeyRepository creates a new MongoDB passkey repository
func NewPasskeyRepository(db *mongo.Database) ports.PasskeyRepository {
	return &passkeyRepository{
		db:               db,
		credentialsColl:  db.Collection("passkey_credentials"),
		userPasskeysColl: db.Collection("user_passkeys"),
	}
}

// Credential operations

func (r *passkeyRepository) CreateCredential(ctx context.Context, credential *domain.PasskeyCredential) error {
	if credential.ID.IsZero() {
		credential.ID = primitive.NewObjectID()
	}
	credential.CreatedAt = time.Now()
	credential.UpdatedAt = time.Now()

	_, err := r.credentialsColl.InsertOne(ctx, credential)
	if err != nil {
		return fmt.Errorf("failed to create passkey credential: %w", err)
	}

	return nil
}

func (r *passkeyRepository) GetCredentialByID(ctx context.Context, id primitive.ObjectID) (*domain.PasskeyCredential, error) {
	var credential domain.PasskeyCredential
	err := r.credentialsColl.FindOne(ctx, bson.M{"_id": id}).Decode(&credential)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get passkey credential: %w", err)
	}

	return &credential, nil
}

func (r *passkeyRepository) GetCredentialByCredentialID(ctx context.Context, credentialID []byte) (*domain.PasskeyCredential, error) {
	var credential domain.PasskeyCredential
	err := r.credentialsColl.FindOne(ctx, bson.M{"credentialId": credentialID}).Decode(&credential)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get passkey credential: %w", err)
	}

	return &credential, nil
}

func (r *passkeyRepository) UpdateCredentialSignCount(ctx context.Context, id primitive.ObjectID, signCount uint32) error {
	update := bson.M{
		"$set": bson.M{
			"signCount": signCount,
			"updatedAt": time.Now(),
		},
	}

	_, err := r.credentialsColl.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("failed to update passkey credential sign count: %w", err)
	}

	return nil
}

func (r *passkeyRepository) DeleteCredential(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.credentialsColl.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete passkey credential: %w", err)
	}

	return nil
}

// User-Passkey relationship operations

func (r *passkeyRepository) AssignCredentialToUser(ctx context.Context, userPasskey *domain.UserPasskey) error {
	if userPasskey.ID.IsZero() {
		userPasskey.ID = primitive.NewObjectID()
	}
	userPasskey.IsActive = true
	userPasskey.CreatedAt = time.Now()
	userPasskey.UpdatedAt = time.Now()
	userPasskey.LastUsedAt = time.Now()

	_, err := r.userPasskeysColl.InsertOne(ctx, userPasskey)
	if err != nil {
		return fmt.Errorf("failed to assign passkey credential to user: %w", err)
	}

	return nil
}

func (r *passkeyRepository) GetUserPasskeys(ctx context.Context, userID primitive.ObjectID) ([]*domain.UserPasskey, error) {
	cursor, err := r.userPasskeysColl.Find(ctx, bson.M{"userId": userID})
	if err != nil {
		return nil, fmt.Errorf("failed to get user passkeys: %w", err)
	}
	defer cursor.Close(ctx)

	var userPasskeys []*domain.UserPasskey
	if err := cursor.All(ctx, &userPasskeys); err != nil {
		return nil, fmt.Errorf("failed to decode user passkeys: %w", err)
	}

	return userPasskeys, nil
}

func (r *passkeyRepository) GetActiveUserPasskeys(ctx context.Context, userID primitive.ObjectID) ([]*domain.UserPasskey, error) {
	cursor, err := r.userPasskeysColl.Find(ctx, bson.M{
		"userId":   userID,
		"isActive": true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get active user passkeys: %w", err)
	}
	defer cursor.Close(ctx)

	var userPasskeys []*domain.UserPasskey
	if err := cursor.All(ctx, &userPasskeys); err != nil {
		return nil, fmt.Errorf("failed to decode active user passkeys: %w", err)
	}

	return userPasskeys, nil
}

func (r *passkeyRepository) DeactivateUserPasskey(ctx context.Context, id primitive.ObjectID) error {
	update := bson.M{
		"$set": bson.M{
			"isActive":  false,
			"updatedAt": time.Now(),
		},
	}

	_, err := r.userPasskeysColl.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("failed to deactivate user passkey: %w", err)
	}

	return nil
}

func (r *passkeyRepository) UpdateUserPasskeyLastUsed(ctx context.Context, id primitive.ObjectID, deviceInfo string) error {
	update := bson.M{
		"$set": bson.M{
			"lastUsedAt": time.Now(),
			"deviceInfo": deviceInfo,
			"updatedAt":  time.Now(),
		},
	}

	_, err := r.userPasskeysColl.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("failed to update user passkey last used: %w", err)
	}

	return nil
}
