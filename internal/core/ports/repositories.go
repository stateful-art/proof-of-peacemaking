package ports

import (
	"context"

	"github.com/stateful-art/proof-of-peacemaking/web/internal/core/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByAddress(ctx context.Context, address string) (*domain.User, error)
	UpdateNonce(ctx context.Context, id primitive.ObjectID, nonce int) error
}

type NotificationRepository interface {
	Create(ctx context.Context, notification *domain.Notification) error
	CreateUserNotification(ctx context.Context, userNotification *domain.UserNotification) error
	GetUserUnreadNotifications(ctx context.Context, userID primitive.ObjectID) ([]*domain.Notification, error)
	MarkAsRead(ctx context.Context, userID, notificationID primitive.ObjectID) error
}
