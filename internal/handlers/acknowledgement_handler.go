package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/stateful-art/proof-of-peacemaking/internal/core/ports"
)

type AcknowledgementHandler struct {
	acknowledgementService ports.AcknowledgementService
}

func NewAcknowledgementHandler(acknowledgementService ports.AcknowledgementService) *AcknowledgementHandler {
	return &AcknowledgementHandler{
		acknowledgementService: acknowledgementService,
	}
}

func (h *AcknowledgementHandler) Create(c *fiber.Ctx) error {
	var body struct {
		ExpressionID string            `json:"expressionId"`
		Content      map[string]string `json:"content"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	userAddress := c.Locals("userAddress").(string)
	acknowledgement, err := h.acknowledgementService.Create(c.Context(), userAddress, body.ExpressionID, body.Content)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create acknowledgement",
		})
	}

	return c.JSON(acknowledgement)
}

func (h *AcknowledgementHandler) ListByExpression(c *fiber.Ctx) error {
	expressionID := c.Params("id")
	acknowledgements, err := h.acknowledgementService.ListByExpression(c.Context(), expressionID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch acknowledgements",
		})
	}
	return c.JSON(acknowledgements)
}
