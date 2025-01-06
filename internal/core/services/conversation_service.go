package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"
	"time"

	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type conversationService struct {
	repo                ports.ConversationRepository
	livekitClient       *lksdk.RoomServiceClient
	notificationService ports.NotificationService
	apiKey              string
	apiSecret           string
}

func NewConversationService(
	repo ports.ConversationRepository,
	livekitClient *lksdk.RoomServiceClient,
	notificationService ports.NotificationService,
) ports.ConversationService {
	return &conversationService{
		repo:                repo,
		livekitClient:       livekitClient,
		notificationService: notificationService,
		apiKey:              os.Getenv("LIVEKIT_API_KEY"),
		apiSecret:           os.Getenv("LIVEKIT_API_SECRET"),
	}
}

func (s *conversationService) CreateConversation(ctx context.Context, conversation *domain.Conversation) error {
	// Generate a unique room name for LiveKit
	conversation.RoomName = fmt.Sprintf("conversation-%s", primitive.NewObjectID().Hex())
	conversation.Status = domain.ConversationStatusScheduled
	conversation.CreatedAt = time.Now()
	conversation.UpdatedAt = time.Now()

	// Create LiveKit room
	_, err := s.livekitClient.CreateRoom(ctx, &livekit.CreateRoomRequest{
		Name:            conversation.RoomName,
		EmptyTimeout:    300, // 5 minutes
		MaxParticipants: 100,
	})
	if err != nil {
		return fmt.Errorf("failed to create LiveKit room: %w", err)
	}

	if err := s.repo.Create(ctx, conversation); err != nil {
		return err
	}

	// Send notification
	notification := domain.Notification{
		UserID:  conversation.CreatorID,
		Type:    domain.NotificationConversationCreated,
		Title:   "New Conversation Created",
		Message: "Your conversation has been created successfully",
		Data: map[string]interface{}{
			"conversationId": conversation.ID,
			"title":          conversation.Title,
		},
		CreatedAt: time.Now(),
	}

	return s.notificationService.CreateNotification(ctx, notification)
}

func (s *conversationService) GetConversation(ctx context.Context, id primitive.ObjectID) (*domain.Conversation, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *conversationService) ListConversations(ctx context.Context, filter map[string]interface{}) ([]*domain.Conversation, error) {
	return s.repo.List(ctx, filter)
}

func (s *conversationService) StartConversation(ctx context.Context, id primitive.ObjectID, userID string) error {
	log.Printf("[DEBUG] Starting conversation service. ID: %s, UserID: %s", id.Hex(), userID)

	conversation, err := s.repo.GetByID(ctx, id)
	if err != nil {
		log.Printf("[ERROR] Failed to get conversation: %v", err)
		return err
	}

	if conversation == nil {
		log.Printf("[ERROR] Conversation not found: %s", id.Hex())
		return fmt.Errorf("conversation not found")
	}

	log.Printf("[DEBUG] Found conversation. CreatorID: %s, Status: %s", conversation.CreatorID, conversation.Status)

	if conversation.CreatorID != userID {
		log.Printf("[ERROR] User %s is not the creator %s", userID, conversation.CreatorID)
		return fmt.Errorf("only the creator can start the conversation")
	}

	if conversation.Status != domain.ConversationStatusScheduled {
		log.Printf("[ERROR] Invalid conversation status: %s", conversation.Status)
		return fmt.Errorf("conversation is not in scheduled status")
	}

	conversation.Status = domain.ConversationStatusLive
	conversation.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, conversation); err != nil {
		log.Printf("[ERROR] Failed to update conversation: %v", err)
		return err
	}

	// Send notification to subscribers
	if conversation.Subscribers != nil {
		for _, subscriberID := range conversation.Subscribers {
			notification := domain.Notification{
				UserID:  subscriberID,
				Type:    domain.NotificationConversationStarted,
				Title:   "Conversation Started",
				Message: fmt.Sprintf("The conversation '%s' has started", conversation.Title),
				Data: map[string]interface{}{
					"conversationId": conversation.ID,
					"title":          conversation.Title,
				},
				CreatedAt: time.Now(),
			}

			if err := s.notificationService.CreateNotification(ctx, notification); err != nil {
				log.Printf("[WARN] Failed to send notification to subscriber %s: %v", subscriberID, err)
			}
		}
	}

	log.Printf("[INFO] Successfully started conversation %s", id.Hex())
	return nil
}

func (s *conversationService) EndConversation(ctx context.Context, id primitive.ObjectID, userID string) error {
	conversation, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if conversation == nil {
		return fmt.Errorf("conversation not found")
	}

	if conversation.CreatorID != userID {
		return fmt.Errorf("only the creator can end the conversation")
	}

	if conversation.Status != domain.ConversationStatusLive {
		return fmt.Errorf("conversation is not live")
	}

	conversation.Status = domain.ConversationStatusEnded
	conversation.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, conversation); err != nil {
		return err
	}

	// Send notification to subscribers
	if conversation.Subscribers != nil {
		for _, subscriberID := range conversation.Subscribers {
			notification := domain.Notification{
				UserID:  subscriberID,
				Type:    domain.NotificationConversationEnded,
				Title:   "Conversation Ended",
				Message: fmt.Sprintf("The conversation '%s' has ended", conversation.Title),
				Data: map[string]interface{}{
					"conversationId": conversation.ID,
					"title":          conversation.Title,
				},
				CreatedAt: time.Now(),
			}

			if err := s.notificationService.CreateNotification(ctx, notification); err != nil {
				log.Printf("Failed to send notification to subscriber %s: %v", subscriberID, err)
			}
		}
	}

	return nil
}

func (s *conversationService) SubscribeToNotifications(ctx context.Context, id primitive.ObjectID, userID string) error {
	conversation, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if conversation == nil {
		return fmt.Errorf("conversation not found")
	}

	return s.repo.AddSubscriber(ctx, id, userID)
}

func (s *conversationService) UnsubscribeFromNotifications(ctx context.Context, id primitive.ObjectID, userID string) error {
	conversation, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if conversation == nil {
		return fmt.Errorf("conversation not found")
	}

	return s.repo.RemoveSubscriber(ctx, id, userID)
}

func (s *conversationService) GenerateJoinToken(userID string, roomName string, canPublish bool) (string, error) {
	// Generate LiveKit token
	at := auth.NewAccessToken(s.apiKey, s.apiSecret)
	at.SetIdentity(userID)

	grant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     roomName,
	}

	// Convert bool to *bool
	canPublishPtr := &canPublish
	canSubscribePtr := new(bool)
	*canSubscribePtr = true

	// Set publish and subscribe permissions
	grant.CanPublish = canPublishPtr
	grant.CanSubscribe = canSubscribePtr

	at.AddGrant(grant)

	return at.ToJWT()
}
