package ports

import (
	"context"

	"proofofpeacemaking/internal/core/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ConversationService interface {
	CreateConversation(ctx context.Context, conversation *domain.Conversation) error
	GetConversation(ctx context.Context, id primitive.ObjectID) (*domain.Conversation, error)
	ListConversations(ctx context.Context, filter map[string]interface{}) ([]*domain.Conversation, error)
	StartConversation(ctx context.Context, id primitive.ObjectID, userID string) error
	EndConversation(ctx context.Context, id primitive.ObjectID, userID string) error
	SubscribeToNotifications(ctx context.Context, id primitive.ObjectID, userID string) error
	UnsubscribeFromNotifications(ctx context.Context, id primitive.ObjectID, userID string) error
	GenerateJoinToken(userID string, roomName string, canPublish bool) (string, error)
}
