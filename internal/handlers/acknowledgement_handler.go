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
	statsService           ports.StatisticsService
}

func NewAcknowledgementHandler(
	acknowledgementService ports.AcknowledgementService,
	userService ports.UserService,
	expressionService ports.ExpressionService,
	statsService ports.StatisticsService,
) *AcknowledgementHandler {
	return &AcknowledgementHandler{
		acknowledgementService: acknowledgementService,
		userService:            userService,
		expressionService:      expressionService,
		statsService:           statsService,
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

	log.Printf("[ACK] Creating/updating acknowledgement for expression: %s", body.ExpressionID)
	userIdentifier := c.Locals("userAddress").(string)
	log.Printf("[ACK] User identifier from context: %s", userIdentifier)

	// Get user by email or address
	var user *domain.User
	var err error
	if strings.Contains(userIdentifier, "@") {
		user, err = h.userService.GetUserByEmail(c.Context(), userIdentifier)
	} else {
		user, err = h.userService.GetUserByAddress(c.Context(), userIdentifier)
	}

	if err != nil {
		log.Printf("[ACK] Error getting user: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user",
		})
	}
	if user == nil {
		log.Printf("[ACK] User not found for identifier: %s", userIdentifier)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}
	log.Printf("[ACK] Found user with ID: %s", user.ID.Hex())

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

	// Prevent self-acknowledgements - compare user IDs instead of addresses
	if expression.Creator == user.ID.Hex() {
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

	var existingAck *domain.Acknowledgement
	for _, ack := range existingAcks {
		if ack.Acknowledger == user.ID.Hex() {
			existingAck = ack
			break
		}
	}

	if existingAck != nil {
		// Update existing acknowledgement status
		if existingAck.Status == domain.AcknowledgementStatusActive {
			existingAck.Status = domain.AcknowledgementStatusRefuted
		} else {
			existingAck.Status = domain.AcknowledgementStatusActive
		}
		existingAck.UpdatedAt = time.Now()

		if err := h.acknowledgementService.Update(c.Context(), existingAck); err != nil {
			log.Printf("[ACK] Error updating acknowledgement: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update acknowledgement",
			})
		}

		return c.JSON(existingAck)
	}

	// Create new acknowledgement
	acknowledgement := &domain.Acknowledgement{
		ID:           primitive.NewObjectID(),
		ExpressionID: body.ExpressionID,
		Acknowledger: user.ID.Hex(),
		Content:      body.Content,
		Status:       domain.AcknowledgementStatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := h.acknowledgementService.Create(c.Context(), acknowledgement); err != nil {
		log.Printf("[ACK] Error creating acknowledgement: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create acknowledgement",
		})
	}

	log.Printf("[ACK] Successfully created acknowledgement: %s", acknowledgement.ID.Hex())

	// After creating/updating acknowledgement, update statistics
	if err := h.statsService.UpdateStatisticsAfterAcknowledgement(c.Context()); err != nil {
		log.Printf("[ACK] Warning: Failed to update statistics: %v", err)
		// Don't return error here, as the acknowledgement was created successfully
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
