package services

import (
	"context"
	"fmt"

	"github.com/stateful-art/proof-of-peacemaking/web/internal/core/domain"
	"github.com/stateful-art/proof-of-peacemaking/web/internal/core/ports"
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

// ... implement other notification methods similarly
