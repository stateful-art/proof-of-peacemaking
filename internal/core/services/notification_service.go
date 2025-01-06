package services

import (
	"context"
	"sync"

	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"
)

type notificationService struct {
	repo        ports.NotificationRepository
	subscribers map[string][]chan domain.Notification
	mu          sync.RWMutex
}

func NewNotificationService(repo ports.NotificationRepository) ports.NotificationService {
	return &notificationService{
		repo:        repo,
		subscribers: make(map[string][]chan domain.Notification),
	}
}

func (s *notificationService) GetUserNotifications(ctx context.Context, userAddress string) ([]domain.Notification, error) {
	return s.repo.GetByUser(ctx, userAddress)
}

func (s *notificationService) MarkNotificationAsRead(ctx context.Context, userAddress string, notificationID string) error {
	return s.repo.MarkAsRead(ctx, userAddress, notificationID)
}

func (s *notificationService) MarkAllNotificationsAsRead(ctx context.Context, userAddress string) error {
	return s.repo.MarkAllAsRead(ctx, userAddress)
}

func (s *notificationService) CreateNotification(ctx context.Context, notification domain.Notification) error {
	if err := s.repo.Create(ctx, notification); err != nil {
		return err
	}

	// Notify subscribers
	s.mu.RLock()
	defer s.mu.RUnlock()

	if channels, ok := s.subscribers[notification.UserID]; ok {
		for _, ch := range channels {
			select {
			case ch <- notification:
			default:
				// Channel is full or closed, skip it
			}
		}
	}

	return nil
}

func (s *notificationService) SubscribeToNotifications(ctx context.Context, userAddress string) (<-chan domain.Notification, error) {
	ch := make(chan domain.Notification, 100)

	s.mu.Lock()
	if _, ok := s.subscribers[userAddress]; !ok {
		s.subscribers[userAddress] = make([]chan domain.Notification, 0)
	}
	s.subscribers[userAddress] = append(s.subscribers[userAddress], ch)
	s.mu.Unlock()

	// Clean up when context is done
	go func() {
		<-ctx.Done()
		s.mu.Lock()
		defer s.mu.Unlock()

		if channels, ok := s.subscribers[userAddress]; ok {
			newChannels := make([]chan domain.Notification, 0)
			for _, existingCh := range channels {
				if existingCh != ch {
					newChannels = append(newChannels, existingCh)
				}
			}
			if len(newChannels) == 0 {
				delete(s.subscribers, userAddress)
			} else {
				s.subscribers[userAddress] = newChannels
			}
		}
		close(ch)
	}()

	return ch, nil
}
