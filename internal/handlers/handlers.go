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
	YouTube         *YouTubeHandler
	Song            *SongHandler
	Conversation    *ConversationHandler
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
	songService ports.SongService,
	conversationService ports.ConversationService,
	notificationService ports.NotificationService,
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
		YouTube:         NewYouTubeHandler(),
		Song:            NewSongHandler(songService),
		Conversation:    NewConversationHandler(conversationService, userService),
		Notification:    NewNotificationHandler(notificationService),
	}
}
