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
	WebAuthn        *WebAuthnHandler
	Statistics      *StatisticsHandler
	Account         *AccountHandler
}

func NewHandlers(
	userService ports.UserService,
	authService ports.AuthService,
	expressionService ports.ExpressionService,
	acknowledgementService ports.AcknowledgementService,
	proofNFTService ports.ProofNFTService,
	feedService ports.FeedService,
	statisticsService ports.StatisticsService,
	webAuthnService ports.WebAuthnService,
	sessionService ports.SessionService,
	newsletterService ports.NewsletterService,
) *Handlers {
	return &Handlers{
		Auth:            NewAuthHandler(authService, userService),
		User:            NewUserHandler(userService, statisticsService),
		Expression:      NewExpressionHandler(expressionService, userService, statisticsService),
		Acknowledgement: NewAcknowledgementHandler(acknowledgementService, userService, expressionService, statisticsService),
		ProofNFT:        NewProofNFTHandler(proofNFTService),
		Feed:            NewFeedHandler(feedService, userService),
		Statistics:      NewStatisticsHandler(statisticsService),
		Account:         NewAccountHandler(userService, authService, statisticsService),
		WebAuthn:        NewWebAuthnHandler(webAuthnService, sessionService, userService),
		Newsletter:      NewNewsletterHandler(newsletterService),
		Dashboard:       NewDashboardHandler(expressionService, acknowledgementService, userService, proofNFTService),
	}
}
