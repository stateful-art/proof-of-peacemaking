package handlers

import "github.com/stateful-art/proof-of-peacemaking/internal/core/ports"

type Handlers struct {
	Auth            *AuthHandler
	Notification    *NotificationHandler
	Expression      *ExpressionHandler
	Acknowledgement *AcknowledgementHandler
	ProofNFT        *ProofNFTHandler
}

func NewHandlers(
	notificationService ports.NotificationService,
	authService ports.AuthService,
	expressionService ports.ExpressionService,
	acknowledgementService ports.AcknowledgementService,
	proofNFTService ports.ProofNFTService,
) *Handlers {
	return &Handlers{
		Auth:            NewAuthHandler(authService),
		Notification:    NewNotificationHandler(notificationService),
		Expression:      NewExpressionHandler(expressionService),
		Acknowledgement: NewAcknowledgementHandler(acknowledgementService),
		ProofNFT:        NewProofNFTHandler(proofNFTService),
	}
}
