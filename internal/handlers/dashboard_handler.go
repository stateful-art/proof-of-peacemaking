package handlers

import (
	"proofofpeacemaking/internal/core/ports"

	"github.com/gofiber/fiber/v2"
)

type DashboardHandler struct {
	expressionService      ports.ExpressionService
	acknowledgementService ports.AcknowledgementService
	userService            ports.UserService
}

func NewDashboardHandler(
	expressionService ports.ExpressionService,
	acknowledgementService ports.AcknowledgementService,
	userService ports.UserService,
) *DashboardHandler {
	return &DashboardHandler{
		expressionService:      expressionService,
		acknowledgementService: acknowledgementService,
		userService:            userService,
	}
}

func (h *DashboardHandler) GetDashboard(c *fiber.Ctx) error {
	userAddress := c.Locals("userAddress").(string)

	// Get user's expressions and acknowledgements
	expressions, err := h.expressionService.ListByUser(c.Context(), userAddress)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch user expressions",
		})
	}

	acknowledgements, err := h.acknowledgementService.ListByUser(c.Context(), userAddress)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch user acknowledgements",
		})
	}

	// Get user details
	user, err := h.userService.GetUserByAddress(c.Context(), userAddress)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch user details",
		})
	}

	// Render dashboard template with user data
	return c.Render("dashboard", fiber.Map{
		"Title":            "Dashboard - Proof of Peacemaking",
		"User":             user,
		"Expressions":      expressions,
		"Acknowledgements": acknowledgements,
	})
}
