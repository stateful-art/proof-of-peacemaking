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

	userNotification := &domain.UserNotification{
		UserID:         expression.Creator,
		NotificationID: notification.ID,
		CreatedAt:      notification.CreatedAt,
	}

	return s.notificationRepo.CreateUserNotification(ctx, userNotification)
}

func (s *notificationService) GetUserNotifications(ctx context.Context, userAddress string) ([]*domain.Notification, error) {
	user, err := s.userRepo.FindByAddress(ctx, userAddress)
	if err != nil {
		return nil, err
	}
	return s.notificationRepo.GetUserUnreadNotifications(ctx, user.ID)
}

func (s *notificationService) MarkNotificationAsRead(ctx context.Context, userAddress string, notificationID string) error {
	user, err := s.userRepo.FindByAddress(ctx, userAddress)
	if err != nil {
		return err
	}
	notifID, err := primitive.ObjectIDFromHex(notificationID)
	if err != nil {
		return err
	}
	return s.notificationRepo.MarkAsRead(ctx, user.ID, notifID)
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

	// Notify both creator and acknowledger
	for _, userID := range []primitive.ObjectID{nft.Expression, nft.Acknowledger} {
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

func (s *notificationService) NotifyProofRequest(ctx context.Context, request *domain.ProofRequest) error {
	notification := &domain.Notification{
		Type:    domain.NotificationProofRequestReceived,
		Title:   "New Proof Request",
		Message: "You have received a request to create a proof NFT",
		Data: map[string]interface{}{
			"expressionId": request.Expression,
			"requestedBy":  request.Acknowledger,
		},
		CreatedAt: request.CreatedAt,
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	userNotification := &domain.UserNotification{
		UserID:         request.Expression, // Notify the expression creator
		NotificationID: notification.ID,
		CreatedAt:      notification.CreatedAt,
	}

	return s.notificationRepo.CreateUserNotification(ctx, userNotification)
}

// ... implement other notification methods similarly
