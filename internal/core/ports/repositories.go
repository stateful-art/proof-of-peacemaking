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
	Delete(ctx context.Context, id primitive.ObjectID) error
	GetTotalCount(ctx context.Context) (int, error)
	GetCitizenshipDistribution(ctx context.Context) (map[string]int, error)
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
	DeleteByUserID(ctx context.Context, userID primitive.ObjectID) error
}

type ExpressionRepository interface {
	Create(ctx context.Context, expression *domain.Expression) error
	FindByID(ctx context.Context, id string) (*domain.Expression, error)
	FindAll(ctx context.Context) ([]*domain.Expression, error)
	FindByCreatorID(ctx context.Context, creatorID string) ([]*domain.Expression, error)
	FindByIDs(ctx context.Context, ids []string) ([]*domain.Expression, error)
	GetByUserID(ctx context.Context, userID string) ([]*domain.Expression, error)
	Update(ctx context.Context, expression *domain.Expression) error
	Delete(ctx context.Context, id string) error
	GetTotalCount(ctx context.Context) (int, error)
	GetTotalAcknowledgements(ctx context.Context) (int, error)
	GetMediaTypeDistribution(ctx context.Context) (map[string]int, error)
}

type AcknowledgementRepository interface {
	Create(ctx context.Context, acknowledgement *domain.Acknowledgement) error
	FindByExpression(ctx context.Context, expressionID string) ([]*domain.Acknowledgement, error)
	FindByAcknowledger(ctx context.Context, acknowledgerID string) ([]*domain.Acknowledgement, error)
	FindByStatus(ctx context.Context, status domain.AcknowledgementStatus) ([]*domain.Acknowledgement, error)
	Update(ctx context.Context, acknowledgement *domain.Acknowledgement) error
}

// PasskeyRepository handles passkey credential storage operations
type PasskeyRepository interface {
	// Credential operations
	CreateCredential(ctx context.Context, credential *domain.PasskeyCredential) error
	GetCredentialByID(ctx context.Context, id primitive.ObjectID) (*domain.PasskeyCredential, error)
	GetCredentialByCredentialID(ctx context.Context, credentialID []byte) (*domain.PasskeyCredential, error)
	UpdateCredentialSignCount(ctx context.Context, id primitive.ObjectID, signCount uint32) error
	DeleteCredential(ctx context.Context, id primitive.ObjectID) error

	// User-Passkey relationship operations
	AssignCredentialToUser(ctx context.Context, userPasskey *domain.UserPasskey) error
	GetUserPasskeys(ctx context.Context, userID primitive.ObjectID) ([]*domain.UserPasskey, error)
	GetActiveUserPasskeys(ctx context.Context, userID primitive.ObjectID) ([]*domain.UserPasskey, error)
	DeactivateUserPasskey(ctx context.Context, id primitive.ObjectID) error
	UpdateUserPasskeyLastUsed(ctx context.Context, id primitive.ObjectID, deviceInfo string) error
}

// StatisticsRepository handles statistics data storage
type StatisticsRepository interface {
	// GetLatest returns the most recent statistics record
	GetLatest(ctx context.Context) (*domain.Statistics, error)

	// Create stores a new statistics record
	Create(ctx context.Context, stats *domain.Statistics) error

	// GetCountryList returns the list of available countries
	GetCountryList(ctx context.Context) ([]domain.CountryInfo, error)
}
