package mongodb

import (
	"context"
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
	_, err := r.db.Collection("sessions").DeleteMany(ctx, bson.M{
		"expiresAt": bson.M{
			"$lt": time.Now(),
		},
	})
	return err
}
