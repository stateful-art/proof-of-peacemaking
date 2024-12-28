package ports

import (
	"context"

	"proofofpeacemaking/internal/core/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByAddress(ctx context.Context, address string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	UpdateNonce(ctx context.Context, id primitive.ObjectID, nonce int) error
}

type NotificationRepository interface {
	Create(ctx context.Context, notification *domain.Notification) error
	CreateUserNotification(ctx context.Context, userNotification *domain.UserNotification) error
	GetUserUnreadNotifications(ctx context.Context, userID primitive.ObjectID) ([]*domain.Notification, error)
	MarkAsRead(ctx context.Context, userID, notificationID primitive.ObjectID) error
}

type SessionRepository interface {
	Create(ctx context.Context, session *domain.Session) error
	FindByToken(ctx context.Context, token string) (*domain.Session, error)
	DeleteByToken(ctx context.Context, token string) error
	DeleteExpired(ctx context.Context) error
}

type ExpressionRepository interface {
	Create(ctx context.Context, expression *domain.Expression) error
	FindByID(ctx context.Context, id string) (*domain.Expression, error)
	FindAll(ctx context.Context) ([]*domain.Expression, error)
	FindByCreator(ctx context.Context, creatorAddress string) ([]*domain.Expression, error)
}

type AcknowledgementRepository interface {
	Create(ctx context.Context, acknowledgement *domain.Acknowledgement) error
	FindByExpression(ctx context.Context, expressionID string) ([]*domain.Acknowledgement, error)
	FindByCreator(ctx context.Context, creatorAddress string) ([]*domain.Acknowledgement, error)
}
