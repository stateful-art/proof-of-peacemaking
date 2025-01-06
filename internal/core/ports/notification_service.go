package ports

import (
	"context"
	"proofofpeacemaking/internal/core/domain"
)

type NotificationService interface {
	GetUserNotifications(ctx context.Context, userAddress string) ([]domain.Notification, error)
	MarkNotificationAsRead(ctx context.Context, userAddress string, notificationID string) error
	MarkAllNotificationsAsRead(ctx context.Context, userAddress string) error
	CreateNotification(ctx context.Context, notification domain.Notification) error
	SubscribeToNotifications(ctx context.Context, userAddress string) (<-chan domain.Notification, error)
}
