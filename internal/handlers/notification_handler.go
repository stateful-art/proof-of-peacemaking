package handlers

import (
	"proofofpeacemaking/internal/core/ports"

	"github.com/gofiber/fiber/v2"
)

type NotificationHandler struct {
	notificationService ports.NotificationService
}

func NewNotificationHandler(notificationService ports.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
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
