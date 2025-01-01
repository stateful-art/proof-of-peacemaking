package handlers

import (
	"encoding/json"
	"proofofpeacemaking/internal/core/ports"

	"github.com/gofiber/fiber/v2"
)

type NewsletterHandler struct {
	newsletterService ports.NewsletterService
}

func NewNewsletterHandler(newsletterService ports.NewsletterService) *NewsletterHandler {
	return &NewsletterHandler{
		newsletterService: newsletterService,
	}
}

func (h *NewsletterHandler) GetNewsletterService() ports.NewsletterService {
	return h.newsletterService
}

func (h *NewsletterHandler) HandleNewsletterRegistration(c *fiber.Ctx) error {
	var newsletterData struct {
		Email string `json:"email"`
	}

	if err := json.Unmarshal(c.Body(), &newsletterData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	if newsletterData.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "All fields are required"})
	}

	err := h.newsletterService.SendContactEmail(c.Context(), newsletterData.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error registering email"})
	}

	return c.SendStatus(fiber.StatusOK)
}
