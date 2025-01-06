package mongodb

import (
	"context"
	"time"

	"proofofpeacemaking/internal/core/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ConversationRepository struct {
	collection *mongo.Collection
}

func NewConversationRepository(db *mongo.Database) *ConversationRepository {
	return &ConversationRepository{
		collection: db.Collection("conversations"),
	}
}

func (r *ConversationRepository) Create(ctx context.Context, conversation *domain.Conversation) error {
	conversation.CreatedAt = time.Now()
	conversation.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, conversation)
	if err != nil {
		return err
	}

	conversation.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *ConversationRepository) Update(ctx context.Context, conversation *domain.Conversation) error {
	conversation.UpdatedAt = time.Now()

	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": conversation.ID}, conversation)
	return err
}

func (r *ConversationRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *ConversationRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Conversation, error) {
	var conversation domain.Conversation
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&conversation)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &conversation, nil
}

func (r *ConversationRepository) List(ctx context.Context, filter map[string]interface{}) ([]*domain.Conversation, error) {
	bsonFilter := bson.M{}
	for k, v := range filter {
		bsonFilter[k] = v
	}

	cursor, err := r.collection.Find(ctx, bsonFilter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var conversations []*domain.Conversation
	if err = cursor.All(ctx, &conversations); err != nil {
		return nil, err
	}

	return conversations, nil
}

func (r *ConversationRepository) AddSubscriber(ctx context.Context, conversationID primitive.ObjectID, userID string) error {
	_, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": conversationID},
		bson.M{"$addToSet": bson.M{"subscribers": userID}})
	return err
}

func (r *ConversationRepository) RemoveSubscriber(ctx context.Context, conversationID primitive.ObjectID, userID string) error {
	_, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": conversationID},
		bson.M{"$pull": bson.M{"subscribers": userID}})
	return err
}

func (r *ConversationRepository) UpdateStatus(ctx context.Context, conversationID primitive.ObjectID, status domain.ConversationStatus) error {
	update := bson.M{
		"$set": bson.M{
			"status":    status,
			"updatedAt": time.Now(),
		},
	}

	if status == domain.ConversationStatusEnded {
		update["$set"].(bson.M)["endTime"] = time.Now()
	}

	_, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": conversationID},
		update)
	return err
}
