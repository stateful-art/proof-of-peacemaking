package handlers

import "proofofpeacemaking/internal/core/ports"

type Handlers struct {
	Notification    *NotificationHandler
	Auth            *AuthHandler
	Expression      *ExpressionHandler
	Acknowledgement *AcknowledgementHandler
	ProofNFT        *ProofNFTHandler
	Feed            *FeedHandler
	Dashboard       *DashboardHandler
}

func NewHandlers(
	notificationService ports.NotificationService,
	authService ports.AuthService,
	expressionService ports.ExpressionService,
	acknowledgementService ports.AcknowledgementService,
	proofNFTService ports.ProofNFTService,
	feedService ports.FeedService,
	userService ports.UserService,
) *Handlers {
	return &Handlers{
		Notification:    NewNotificationHandler(notificationService),
		Auth:            NewAuthHandler(authService),
		Expression:      NewExpressionHandler(expressionService, userService),
		Acknowledgement: NewAcknowledgementHandler(acknowledgementService, userService),
		ProofNFT:        NewProofNFTHandler(proofNFTService),
		Feed:            NewFeedHandler(feedService),
		Dashboard:       NewDashboardHandler(expressionService, acknowledgementService, userService, proofNFTService),
	}
}
