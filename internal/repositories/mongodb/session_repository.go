package mongodb

import (
	"context"
	"fmt"
	"proofofpeacemaking/internal/core/domain"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type sessionRepository struct {
	db *mongo.Database
}

func NewSessionRepository(db *mongo.Database) *sessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) Create(ctx context.Context, session *domain.Session) error {
	_, err := r.db.Collection("sessions").InsertOne(ctx, session)
	return err
}

func (r *sessionRepository) FindByToken(ctx context.Context, token string) (*domain.Session, error) {
	var session domain.Session
	err := r.db.Collection("sessions").FindOne(ctx, bson.M{
		"token": token,
		"expiresAt": bson.M{
			"$gt": time.Now(),
		},
	}).Decode(&session)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &session, err
}

func (r *sessionRepository) DeleteByToken(ctx context.Context, token string) error {
	_, err := r.db.Collection("sessions").DeleteOne(ctx, bson.M{"token": token})
	return err
}

func (r *sessionRepository) DeleteExpired(ctx context.Context) error {
	filter := bson.M{"expiresAt": bson.M{"$lt": time.Now()}}
	_, err := r.db.Collection("sessions").DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}
	return nil
}

func (r *sessionRepository) Update(ctx context.Context, session *domain.Session) error {
	filter := bson.M{"_id": session.ID}
	update := bson.M{
		"$set": bson.M{
			"expiresAt": session.ExpiresAt,
			"updatedAt": session.UpdatedAt,
		},
	}

	result, err := r.db.Collection("sessions").UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("session not found")
	}

	return nil
}
