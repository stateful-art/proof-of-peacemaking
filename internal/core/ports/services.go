package ports

import (
	"context"
	"io"

	"proofofpeacemaking/internal/core/domain"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
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
	RegisterWithEmail(ctx context.Context, email string, password string, username string) (*domain.User, string, error)
	LoginWithEmail(ctx context.Context, email string, password string) (*domain.User, string, error)
	VerifyToken(ctx context.Context, token string) (string, error)
	Logout(ctx context.Context, token string) error
	DeleteAllUserSessions(ctx context.Context, userIdentifier string) error
}

type ExpressionService interface {
	Create(ctx context.Context, expression *domain.Expression) error
	Get(ctx context.Context, id string) (*domain.Expression, error)
	List(ctx context.Context) ([]*domain.Expression, error)
	ListByUser(ctx context.Context, userID string) ([]*domain.Expression, error)
	GetMultiple(ctx context.Context, ids []string) (map[string]*domain.Expression, error)
	UploadMedia(ctx context.Context, expressionID string, mediaType string, reader io.Reader, filename string) (string, error)
	GetMedia(ctx context.Context, expressionID string, mediaType string) (io.ReadCloser, error)
	DeleteMedia(ctx context.Context, expressionID string, mediaType string) error
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
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	GetUserByAddress(ctx context.Context, address string) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByUsername(ctx context.Context, username string) (*domain.User, error)
	UpdateNonce(ctx context.Context, id primitive.ObjectID, nonce int) error
	ConnectWallet(ctx context.Context, userID primitive.ObjectID, address string) error
	Delete(ctx context.Context, id primitive.ObjectID) error
}

type NewsletterService interface {
	SendContactEmail(ctx context.Context, who string) error
}

// WebAuthnService handles WebAuthn operations
type WebAuthnService interface {
	BeginRegistration(ctx context.Context, userID primitive.ObjectID) (*protocol.CredentialCreation, webauthn.SessionData, error)
	FinishRegistration(ctx context.Context, userID primitive.ObjectID, sessionData webauthn.SessionData, response *protocol.ParsedCredentialCreationData) error
	BeginAuthentication(ctx context.Context, userID primitive.ObjectID) (*protocol.CredentialAssertion, webauthn.SessionData, error)
	FinishAuthentication(ctx context.Context, userID primitive.ObjectID, sessionData webauthn.SessionData, response *protocol.ParsedCredentialAssertionData) error
}

// SessionService handles session management
type SessionService interface {
	Create(ctx context.Context, session *domain.Session) error
	GetSession(ctx context.Context, token string) (*domain.Session, error)
	Update(ctx context.Context, session *domain.Session) error
	Delete(ctx context.Context, token string) error
}

// StatisticsService handles system statistics
type StatisticsService interface {
	// GetLatestStats returns the most recent statistics
	GetLatestStats(ctx context.Context) (*domain.Statistics, error)

	// UpdateStats creates a new statistics record
	UpdateStats(ctx context.Context) error

	// GetCountryList returns available countries for citizenship
	GetCountryList(ctx context.Context) ([]domain.CountryInfo, error)

	UpdateStatisticsAfterExpression(ctx context.Context) error
	UpdateStatisticsAfterAcknowledgement(ctx context.Context) error
	UpdateStatisticsAfterCitizenshipChange(ctx context.Context) error
}

// CountryService handles country-related operations
type CountryService interface {
	SearchCountries(ctx context.Context, query string) ([]string, error)
}
