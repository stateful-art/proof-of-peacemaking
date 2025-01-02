package ports

import (
	"context"

	"proofofpeacemaking/internal/core/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByAddress(ctx context.Context, address string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	UpdateNonce(ctx context.Context, id primitive.ObjectID, nonce int) error
	ConnectWallet(ctx context.Context, userID primitive.ObjectID, address string) error
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
	Update(ctx context.Context, session *domain.Session) error
}

type ExpressionRepository interface {
	Create(ctx context.Context, expression *domain.Expression) error
	FindByID(ctx context.Context, id string) (*domain.Expression, error)
	FindAll(ctx context.Context) ([]*domain.Expression, error)
	FindByCreatorID(ctx context.Context, creatorID string) ([]*domain.Expression, error)
	FindByIDs(ctx context.Context, ids []string) ([]*domain.Expression, error)
}

type AcknowledgementRepository interface {
	Create(ctx context.Context, acknowledgement *domain.Acknowledgement) error
	FindByExpression(ctx context.Context, expressionID string) ([]*domain.Acknowledgement, error)
	FindByAcknowledger(ctx context.Context, acknowledgerID string) ([]*domain.Acknowledgement, error)
	FindByStatus(ctx context.Context, status domain.AcknowledgementStatus) ([]*domain.Acknowledgement, error)
	Update(ctx context.Context, acknowledgement *domain.Acknowledgement) error
}
