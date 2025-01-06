package ports

import (
	"context"
	"proofofpeacemaking/internal/core/domain"
)

type NotificationRepository interface {
	Create(ctx context.Context, notification domain.Notification) error
	GetByUser(ctx context.Context, userID string) ([]domain.Notification, error)
	MarkAsRead(ctx context.Context, userID string, notificationID string) error
	MarkAllAsRead(ctx context.Context, userID string) error
}
