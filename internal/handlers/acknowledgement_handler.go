package handlers

import (
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AcknowledgementHandler struct {
	acknowledgementService ports.AcknowledgementService
	userService            ports.UserService
}

func NewAcknowledgementHandler(acknowledgementService ports.AcknowledgementService, userService ports.UserService) *AcknowledgementHandler {
	return &AcknowledgementHandler{
		acknowledgementService: acknowledgementService,
		userService:            userService,
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

	// Get user from address
	user, err := h.userService.GetUserByAddress(c.Context(), userAddress)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user",
		})
	}

	// Convert expression ID to ObjectID
	expressionID, err := primitive.ObjectIDFromHex(body.ExpressionID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid expression ID",
		})
	}

	// Create acknowledgement domain object
	acknowledgement := &domain.Acknowledgement{
		ExpressionID: expressionID,
		Acknowledger: user.ID,
		Content:      body.Content,
		Status:       "pending",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Call service to create acknowledgement
	if err := h.acknowledgementService.Create(c.Context(), acknowledgement); err != nil {
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
