package handlers

import (
	"context"
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

type NotificationHandler struct {
	notificationService ports.NotificationService
}

func NewNotificationHandler(notificationService ports.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

func (h *NotificationHandler) GetNotificationService() ports.NotificationService {
	return h.notificationService
}

func (h *NotificationHandler) HandleWebSocket(c *websocket.Conn) {
	defer c.Close()

	// Get user address from locals
	userAddress, ok := c.Locals("userAddress").(string)
	if !ok {
		return
	}

	// Subscribe to notifications
	notificationChan, err := h.notificationService.SubscribeToNotifications(context.Background(), userAddress)
	if err != nil {
		return
	}

	// Send notifications over WebSocket
	for notification := range notificationChan {
		if err := c.WriteJSON(notification); err != nil {
			break
		}
	}
}

func (h *NotificationHandler) GetUserNotifications(c *fiber.Ctx) error {
	userAddress := c.Locals("userAddress").(string)

	notifications, err := h.notificationService.GetUserNotifications(c.Context(), userAddress)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch notifications",
		})
	}

	return c.JSON(notifications)
}

func (h *NotificationHandler) MarkAsRead(c *fiber.Ctx) error {
	userAddress := c.Locals("userAddress").(string)
	notificationID := c.Params("id")

	err := h.notificationService.MarkNotificationAsRead(c.Context(), userAddress, notificationID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to mark notification as read",
		})
	}

	return c.SendStatus(fiber.StatusOK)
}

func (h *NotificationHandler) MarkAllAsRead(c *fiber.Ctx) error {
	userAddress := c.Locals("userAddress").(string)

	err := h.notificationService.MarkAllNotificationsAsRead(c.Context(), userAddress)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to mark all notifications as read",
		})
	}

	return c.SendStatus(fiber.StatusOK)
}

func (h *NotificationHandler) SubscribeToNotifications(ctx fiber.Ctx, userAddress string) (<-chan domain.Notification, error) {
	return h.notificationService.SubscribeToNotifications(ctx.Context(), userAddress)
}
