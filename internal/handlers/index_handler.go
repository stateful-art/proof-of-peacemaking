package handlers

import (
	"proofofpeacemaking/internal/core/ports"

	"github.com/gofiber/fiber/v2"
)

type IndexHandler struct {
	statsService ports.StatisticsService
}

func NewIndexHandler(statsService ports.StatisticsService) *IndexHandler {
	return &IndexHandler{
		statsService: statsService,
	}
}

// ServeIndexPage renders the home page
func (h *IndexHandler) ServeIndexPage(c *fiber.Ctx) error {
	// Get latest statistics for the home page
	stats, err := h.statsService.GetLatestStats(c.Context())
	if err != nil {
		// If we can't get stats, just log it but don't fail the page load
		stats = nil
	}

	// Get user info if logged in
	userIdentifier, ok := c.Locals("userAddress").(string)
	var userData fiber.Map
	if ok && userIdentifier != "" {
		userData = fiber.Map{
			"Email":   userIdentifier,
			"Address": userIdentifier,
		}
	}

	return c.Render("index", fiber.Map{
		"Title":      "Proof of Peacemaking",
		"User":       userData,
		"Statistics": stats,
	})
}
