package handlers

import "proofofpeacemaking/internal/core/ports"

type Handlers struct {
	Auth            *AuthHandler
	User            *UserHandler
	Expression      *ExpressionHandler
	Acknowledgement *AcknowledgementHandler
	ProofNFT        *ProofNFTHandler
	Feed            *FeedHandler
	Dashboard       *DashboardHandler
	Notification    *NotificationHandler
	Newsletter      *NewsletterHandler
}

func NewHandlers(
	notificationService ports.NotificationService,
	authService ports.AuthService,
	expressionService ports.ExpressionService,
	acknowledgementService ports.AcknowledgementService,
	proofNFTService ports.ProofNFTService,
	feedService ports.FeedService,
	userService ports.UserService,
	newsletterService ports.NewsletterService,
) *Handlers {
	return &Handlers{
		Auth:            NewAuthHandler(authService),
		User:            NewUserHandler(userService),
		Expression:      NewExpressionHandler(expressionService, userService),
		Acknowledgement: NewAcknowledgementHandler(acknowledgementService, userService, expressionService),
		ProofNFT:        NewProofNFTHandler(proofNFTService),
		Feed:            NewFeedHandler(feedService, userService),
		Dashboard:       NewDashboardHandler(expressionService, acknowledgementService, userService, proofNFTService),
		Notification:    NewNotificationHandler(notificationService),
		Newsletter:      NewNewsletterHandler(newsletterService),
	}
}
