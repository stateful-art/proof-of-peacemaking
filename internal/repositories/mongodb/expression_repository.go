package mongodb

import (
	"context"
	"fmt"
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
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

func (r *expressionRepository) FindByCreatorID(ctx context.Context, creatorID string) ([]*domain.Expression, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"creator": creatorID})
	if err != nil {
		return nil, fmt.Errorf("failed to find expressions by creator ID: %w", err)
	}
	defer cursor.Close(ctx)

	var expressions []*domain.Expression
	if err := cursor.All(ctx, &expressions); err != nil {
		return nil, fmt.Errorf("failed to decode expressions: %w", err)
	}

	return expressions, nil
}

func (r *expressionRepository) FindByIDs(ctx context.Context, ids []string) ([]*domain.Expression, error) {
	// Convert string IDs to ObjectIDs
	objectIDs := make([]primitive.ObjectID, 0, len(ids))
	for _, id := range ids {
		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, fmt.Errorf("invalid expression ID format: %w", err)
		}
		objectIDs = append(objectIDs, objectID)
	}

	// Query expressions with $in operator
	cursor, err := r.collection.Find(ctx, bson.M{"_id": bson.M{"$in": objectIDs}})
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
