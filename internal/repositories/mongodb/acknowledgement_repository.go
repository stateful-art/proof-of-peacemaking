package mongodb

import (
	"context"
	"fmt"
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type acknowledgementRepository struct {
	collection *mongo.Collection
}

func NewAcknowledgementRepository(db *mongo.Database) ports.AcknowledgementRepository {
	return &acknowledgementRepository{
		collection: db.Collection("acknowledgements"),
	}
}

func (r *acknowledgementRepository) Create(ctx context.Context, acknowledgement *domain.Acknowledgement) error {
	_, err := r.collection.InsertOne(ctx, acknowledgement)
	if err != nil {
		return fmt.Errorf("failed to create acknowledgement: %w", err)
	}
	return nil
}

func (r *acknowledgementRepository) FindByExpression(ctx context.Context, expressionID string) ([]*domain.Acknowledgement, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"expressionId": expressionID})
	if err != nil {
		return nil, fmt.Errorf("failed to find acknowledgements by expression: %w", err)
	}
	defer cursor.Close(ctx)

	var acknowledgements []*domain.Acknowledgement
	if err := cursor.All(ctx, &acknowledgements); err != nil {
		return nil, fmt.Errorf("failed to decode acknowledgements: %w", err)
	}

	return acknowledgements, nil
}

func (r *acknowledgementRepository) FindByCreator(ctx context.Context, creatorAddress string) ([]*domain.Acknowledgement, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"creatorAddress": creatorAddress})
	if err != nil {
		return nil, fmt.Errorf("failed to find acknowledgements by creator: %w", err)
	}
	defer cursor.Close(ctx)

	var acknowledgements []*domain.Acknowledgement
	if err := cursor.All(ctx, &acknowledgements); err != nil {
		return nil, fmt.Errorf("failed to decode acknowledgements: %w", err)
	}

	return acknowledgements, nil
}
