package services

import (
	"context"
	"fmt"

	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type notificationService struct {
	notificationRepo ports.NotificationRepository
	userRepo         ports.UserRepository
}

func NewNotificationService(
	notificationRepo ports.NotificationRepository,
	userRepo ports.UserRepository,
) ports.NotificationService {
	return &notificationService{
		notificationRepo: notificationRepo,
		userRepo:         userRepo,
	}
}

func (s *notificationService) NotifyNewAcknowledgement(
	ctx context.Context,
	expression *domain.Expression,
	acknowledgement *domain.Acknowledgement,
) error {
	notification := &domain.Notification{
		Type:    domain.NotificationNewAcknowledgement,
		Title:   "New Acknowledgement",
		Message: fmt.Sprintf("Your expression received a new acknowledgement"),
		Data: map[string]interface{}{
			"expressionId":      expression.ID,
			"acknowledgementId": acknowledgement.ID,
			"acknowledger":      acknowledgement.Acknowledger,
		},
		CreatedAt: acknowledgement.CreatedAt,
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	creatorID, err := primitive.ObjectIDFromHex(expression.Creator)
	if err != nil {
		return fmt.Errorf("invalid creator ID format: %w", err)
	}

	userNotification := &domain.UserNotification{
		UserID:         creatorID,
		NotificationID: notification.ID,
		CreatedAt:      notification.CreatedAt,
	}

	return s.notificationRepo.CreateUserNotification(ctx, userNotification)
}

func (s *notificationService) GetUserNotifications(ctx context.Context, userAddress string) ([]*domain.Notification, error) {
	user, err := s.userRepo.GetByAddress(ctx, userAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	notifications, err := s.notificationRepo.GetUserUnreadNotifications(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}

	return notifications, nil
}

func (s *notificationService) MarkNotificationAsRead(ctx context.Context, userAddress string, notificationID string) error {
	user, err := s.userRepo.GetByAddress(ctx, userAddress)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	notificationObjectID, err := primitive.ObjectIDFromHex(notificationID)
	if err != nil {
		return fmt.Errorf("invalid notification ID: %w", err)
	}

	err = s.notificationRepo.MarkAsRead(ctx, user.ID, notificationObjectID)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	return nil
}

func (s *notificationService) NotifyNFTMinted(ctx context.Context, nft *domain.ProofNFT) error {
	notification := &domain.Notification{
		Type:    domain.NotificationNFTMinted,
		Title:   "NFT Minted",
		Message: "Your proof of peacemaking NFT has been minted",
		Data: map[string]interface{}{
			"tokenId":      nft.TokenID,
			"expressionId": nft.Expression,
			"ipfsHash":     nft.IPFSHash,
		},
		CreatedAt: nft.CreatedAt,
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	// Convert string IDs to ObjectIDs
	expressionID, err := primitive.ObjectIDFromHex(nft.Expression)
	if err != nil {
		return fmt.Errorf("invalid expression ID format: %w", err)
	}

	acknowledgerID, err := primitive.ObjectIDFromHex(nft.Acknowledger)
	if err != nil {
		return fmt.Errorf("invalid acknowledger ID format: %w", err)
	}

	// Notify both creator and acknowledger
	for _, userID := range []primitive.ObjectID{expressionID, acknowledgerID} {
		userNotification := &domain.UserNotification{
			UserID:         userID,
			NotificationID: notification.ID,
			CreatedAt:      notification.CreatedAt,
		}
		if err := s.notificationRepo.CreateUserNotification(ctx, userNotification); err != nil {
			return err
		}
	}

	return nil
}

func (s *notificationService) NotifyProofRequestReceived(ctx context.Context, request *domain.ProofRequest) error {
	notification := &domain.Notification{
		Type:    domain.NotificationProofRequestReceived,
		Title:   "New Proof Request",
		Message: "You have received a request to create a proof NFT",
		Data: map[string]interface{}{
			"expressionId": request.ExpressionID,
			"requestedBy":  request.InitiatorID,
		},
		CreatedAt: request.CreatedAt,
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	peerID, err := primitive.ObjectIDFromHex(request.PeerID)
	if err != nil {
		return err
	}

	userNotification := &domain.UserNotification{
		UserID:         peerID,
		NotificationID: notification.ID,
		CreatedAt:      notification.CreatedAt,
	}

	return s.notificationRepo.CreateUserNotification(ctx, userNotification)
}

func (s *notificationService) NotifyProofRequestAccepted(ctx context.Context, request *domain.ProofRequest) error {
	notification := &domain.Notification{
		Type:    domain.NotificationProofRequestAccepted,
		Title:   "Proof Request Accepted",
		Message: "Your proof request has been accepted",
		Data: map[string]interface{}{
			"expressionId": request.ExpressionID,
			"requestedBy":  request.InitiatorID,
		},
		CreatedAt: request.CreatedAt,
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	initiatorID, err := primitive.ObjectIDFromHex(request.InitiatorID)
	if err != nil {
		return err
	}

	userNotification := &domain.UserNotification{
		UserID:         initiatorID,
		NotificationID: notification.ID,
		CreatedAt:      notification.CreatedAt,
	}

	return s.notificationRepo.CreateUserNotification(ctx, userNotification)
}

func (s *notificationService) NotifyProofRequestRejected(ctx context.Context, request *domain.ProofRequest) error {
	notification := &domain.Notification{
		Type:    domain.NotificationProofRequestRejected,
		Title:   "Proof Request Rejected",
		Message: "Your proof request has been rejected",
		Data: map[string]interface{}{
			"expressionId": request.ExpressionID,
			"requestedBy":  request.InitiatorID,
		},
		CreatedAt: request.CreatedAt,
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	initiatorID, err := primitive.ObjectIDFromHex(request.InitiatorID)
	if err != nil {
		return err
	}

	userNotification := &domain.UserNotification{
		UserID:         initiatorID,
		NotificationID: notification.ID,
		CreatedAt:      notification.CreatedAt,
	}

	return s.notificationRepo.CreateUserNotification(ctx, userNotification)
}
