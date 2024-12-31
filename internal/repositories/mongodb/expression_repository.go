package mongodb

import (
	"context"
	"fmt"
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type expressionRepository struct {
	collection *mongo.Collection
}

func NewExpressionRepository(db *mongo.Database) ports.ExpressionRepository {
	return &expressionRepository{
		collection: db.Collection("expressions"),
	}
}

func (r *expressionRepository) Create(ctx context.Context, expression *domain.Expression) error {
	_, err := r.collection.InsertOne(ctx, expression)
	if err != nil {
		return fmt.Errorf("failed to create expression: %w", err)
	}
	return nil
}

func (r *expressionRepository) FindByID(ctx context.Context, id string) (*domain.Expression, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid expression ID format: %w", err)
	}

	var expression domain.Expression
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&expression)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find expression: %w", err)
	}
	return &expression, nil
}

func (r *expressionRepository) FindAll(ctx context.Context) ([]*domain.Expression, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to find expressions: %w", err)
	}
	defer cursor.Close(ctx)

	var expressions []*domain.Expression
	if err := cursor.All(ctx, &expressions); err != nil {
		return nil, fmt.Errorf("failed to decode expressions: %w", err)
	}

	return expressions, nil
}

func (r *expressionRepository) FindByCreator(ctx context.Context, creatorAddress string) ([]*domain.Expression, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"creatorAddress": creatorAddress})
	if err != nil {
		return nil, fmt.Errorf("failed to find expressions by creator: %w", err)
	}
	defer cursor.Close(ctx)

	var expressions []*domain.Expression
	if err := cursor.All(ctx, &expressions); err != nil {
		return nil, fmt.Errorf("failed to decode expressions: %w", err)
	}

	return expressions, nil
}
