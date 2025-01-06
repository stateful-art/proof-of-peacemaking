package ports

import (
	"context"
	"proofofpeacemaking/internal/core/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ConversationRepository interface {
	Create(ctx context.Context, conversation *domain.Conversation) error
	Update(ctx context.Context, conversation *domain.Conversation) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Conversation, error)
	List(ctx context.Context, filter map[string]interface{}) ([]*domain.Conversation, error)
	AddSubscriber(ctx context.Context, conversationID primitive.ObjectID, userID string) error
	RemoveSubscriber(ctx context.Context, conversationID primitive.ObjectID, userID string) error
	UpdateStatus(ctx context.Context, conversationID primitive.ObjectID, status domain.ConversationStatus) error
}
