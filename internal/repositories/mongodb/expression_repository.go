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

// Delete removes an expression by its ID
func (r *expressionRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid expression ID format: %w", err)
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return fmt.Errorf("failed to delete expression: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("expression not found")
	}

	return nil
}

// GetByUserID returns all expressions created by a specific user
func (r *expressionRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Expression, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	cursor, err := r.collection.Find(ctx, bson.M{"creator": objectID})
	if err != nil {
		return nil, fmt.Errorf("failed to find expressions by user ID: %w", err)
	}
	defer cursor.Close(ctx)

	var expressions []*domain.Expression
	if err := cursor.All(ctx, &expressions); err != nil {
		return nil, fmt.Errorf("failed to decode expressions: %w", err)
	}

	return expressions, nil
}

// GetTotalCount returns the total number of expressions
func (r *expressionRepository) GetTotalCount(ctx context.Context) (int, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, fmt.Errorf("failed to get total expression count: %w", err)
	}
	return int(count), nil
}

// GetTotalAcknowledgements returns the total number of acknowledgements across all expressions
func (r *expressionRepository) GetTotalAcknowledgements(ctx context.Context) (int, error) {
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id": nil,
				"total": bson.M{
					"$sum": "$acknowledgementCount",
				},
			},
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, fmt.Errorf("failed to get total acknowledgements: %w", err)
	}
	defer cursor.Close(ctx)

	var result struct {
		Total int `bson:"total"`
	}

	if !cursor.Next(ctx) {
		return 0, nil // No documents found
	}

	if err := cursor.Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode total acknowledgements: %w", err)
	}

	return result.Total, nil
}

// GetMediaTypeDistribution returns a map of media types to their counts
func (r *expressionRepository) GetMediaTypeDistribution(ctx context.Context) (map[string]int, error) {
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id": "$mediaType",
				"count": bson.M{
					"$sum": 1,
				},
			},
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to get media type distribution: %w", err)
	}
	defer cursor.Close(ctx)

	var results []struct {
		MediaType string `bson:"_id"`
		Count     int    `bson:"count"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode media type distribution: %w", err)
	}

	distribution := make(map[string]int)
	for _, result := range results {
		if result.MediaType != "" { // Skip empty media types
			distribution[result.MediaType] = result.Count
		}
	}

	return distribution, nil
}

// Update updates an existing expression
func (r *expressionRepository) Update(ctx context.Context, expression *domain.Expression) error {
	filter := bson.M{"_id": expression.ID}
	update := bson.M{"$set": expression}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update expression: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("expression not found")
	}

	return nil
}
