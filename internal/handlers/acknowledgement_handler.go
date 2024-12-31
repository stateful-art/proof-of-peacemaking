package handlers

import (
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"
	"strings"
	"time"

	"log"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AcknowledgementHandler struct {
	acknowledgementService ports.AcknowledgementService
	userService            ports.UserService
	expressionService      ports.ExpressionService
}

func NewAcknowledgementHandler(acknowledgementService ports.AcknowledgementService, userService ports.UserService, expressionService ports.ExpressionService) *AcknowledgementHandler {
	return &AcknowledgementHandler{
		acknowledgementService: acknowledgementService,
		userService:            userService,
		expressionService:      expressionService,
	}
}

func (h *AcknowledgementHandler) Create(c *fiber.Ctx) error {
	var body struct {
		ExpressionID string            `json:"expressionId"`
		Content      map[string]string `json:"content"`
	}

	if err := c.BodyParser(&body); err != nil {
		log.Printf("[ACK] Error parsing body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	log.Printf("[ACK] Creating acknowledgement for expression: %s", body.ExpressionID)
	userAddress := c.Locals("userAddress").(string)
	log.Printf("[ACK] User address from context: %s", userAddress)

	// Get user from address
	user, err := h.userService.GetUserByAddress(c.Context(), userAddress)
	if err != nil {
		log.Printf("[ACK] Error getting user: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user",
		})
	}
	if user == nil {
		log.Printf("[ACK] User not found for address: %s", userAddress)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}
	log.Printf("[ACK] Found user with ID: %s", user.ID.Hex())

	// Convert expression ID to ObjectID
	expressionID, err := primitive.ObjectIDFromHex(body.ExpressionID)
	if err != nil {
		log.Printf("[ACK] Invalid expression ID format: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid expression ID",
		})
	}

	// Get the expression to check ownership
	expression, err := h.expressionService.Get(c.Context(), body.ExpressionID)
	if err != nil {
		if strings.Contains(err.Error(), "invalid expression ID format") {
			log.Printf("[ACK] Invalid expression ID format: %v", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid expression ID format",
			})
		}
		log.Printf("[ACK] Error getting expression: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get expression",
		})
	}
	if expression == nil {
		log.Printf("[ACK] Expression not found: %s", body.ExpressionID)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Expression not found",
		})
	}
	log.Printf("[ACK] Found expression with creator: %s", expression.CreatorAddress)

	// Prevent self-acknowledgements
	if expression.CreatorAddress == userAddress {
		log.Printf("[ACK] Attempted self-acknowledgement")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot acknowledge your own expression",
		})
	}

	// Check if user has already acknowledged this expression
	existingAcks, err := h.acknowledgementService.ListByExpression(c.Context(), body.ExpressionID)
	if err != nil {
		log.Printf("[ACK] Error checking existing acknowledgements: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check existing acknowledgements",
		})
	}

	for _, ack := range existingAcks {
		if ack.Acknowledger == user.ID {
			log.Printf("[ACK] User has already acknowledged this expression")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "You have already acknowledged this expression",
			})
		}
	}

	// Create acknowledgement domain object
	acknowledgement := &domain.Acknowledgement{
		ID:           primitive.NewObjectID(),
		ExpressionID: expressionID,
		Acknowledger: user.ID,
		Content:      body.Content,
		Status:       "pending",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Call service to create acknowledgement
	if err := h.acknowledgementService.Create(c.Context(), acknowledgement); err != nil {
		log.Printf("[ACK] Error creating acknowledgement: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create acknowledgement",
		})
	}

	log.Printf("[ACK] Successfully created acknowledgement: %s", acknowledgement.ID.Hex())
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
