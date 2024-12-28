package handlers

import (
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"
	"time"

	"github.com/gofiber/fiber/v2"
)

type ExpressionHandler struct {
	expressionService ports.ExpressionService
	userService       ports.UserService
}

func NewExpressionHandler(expressionService ports.ExpressionService, userService ports.UserService) *ExpressionHandler {
	return &ExpressionHandler{
		expressionService: expressionService,
		userService:       userService,
	}
}

func (h *ExpressionHandler) Create(c *fiber.Ctx) error {
	var body struct {
		Content map[string]string `json:"content"` // text, audio, video, image
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

	// Create expression domain object
	expression := &domain.Expression{
		Creator:   user.ID,
		Content:   body.Content,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Call service to create expression
	if err := h.expressionService.Create(c.Context(), expression); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create expression",
		})
	}

	return c.JSON(expression)
}

func (h *ExpressionHandler) List(c *fiber.Ctx) error {
	expressions, err := h.expressionService.List(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch expressions",
		})
	}
	return c.JSON(expressions)
}

func (h *ExpressionHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	expression, err := h.expressionService.Get(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch expression",
		})
	}
	return c.JSON(expression)
}
