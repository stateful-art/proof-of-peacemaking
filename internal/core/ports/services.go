package ports

import (
	"context"

	"proofofpeacemaking/internal/core/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificationService interface {
	NotifyNewAcknowledgement(ctx context.Context, expression *domain.Expression, acknowledgement *domain.Acknowledgement) error
	NotifyProofRequestReceived(ctx context.Context, request *domain.ProofRequest) error
	NotifyProofRequestAccepted(ctx context.Context, request *domain.ProofRequest) error
	NotifyProofRequestRejected(ctx context.Context, request *domain.ProofRequest) error
	NotifyNFTMinted(ctx context.Context, nft *domain.ProofNFT) error
	GetUserNotifications(ctx context.Context, userAddress string) ([]*domain.Notification, error)
	MarkNotificationAsRead(ctx context.Context, userAddress string, notificationID string) error
}

type AuthService interface {
	GenerateNonce(ctx context.Context, address string) (int, error)
	VerifySignature(ctx context.Context, address string, signature string) (bool, string, error)
	Register(ctx context.Context, address string, email string) (*domain.User, string, error)
	VerifyToken(ctx context.Context, token string) (string, error)
	Logout(ctx context.Context, token string) error
}

type ExpressionService interface {
	Create(ctx context.Context, expression *domain.Expression) error
	Get(ctx context.Context, id string) (*domain.Expression, error)
	List(ctx context.Context) ([]*domain.Expression, error)
	ListByUser(ctx context.Context, userAddress string) ([]*domain.Expression, error)
}

type AcknowledgementService interface {
	Create(ctx context.Context, acknowledgement *domain.Acknowledgement) error
	ListByExpression(ctx context.Context, expressionID string) ([]*domain.Acknowledgement, error)
	ListByUser(ctx context.Context, userAddress string) ([]*domain.Acknowledgement, error)
	ListByStatus(ctx context.Context, status domain.AcknowledgementStatus) ([]*domain.Acknowledgement, error)
	Update(ctx context.Context, acknowledgement *domain.Acknowledgement) error
}

type ProofNFTService interface {
	RequestProof(ctx context.Context, expressionID string, acknowledgementID string) error
	ApproveProof(ctx context.Context, requestID string) error
	ListUserProofs(ctx context.Context, userAddress string) ([]*domain.ProofNFT, error)
}

// FeedService handles feed-related operations
type FeedService interface {
	GetFeed(ctx context.Context) ([]map[string]interface{}, error)
}

// UserService handles user-related operations
type UserService interface {
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	GetUserByAddress(ctx context.Context, address string) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByUsername(ctx context.Context, username string) (*domain.User, error)
	UpdateNonce(ctx context.Context, id primitive.ObjectID, nonce int) error
	ConnectWallet(ctx context.Context, userID primitive.ObjectID, address string) error
}

type NewsletterService interface {
	SendContactEmail(ctx context.Context, who string) error
}
