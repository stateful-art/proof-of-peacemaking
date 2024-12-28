package handlers

import (
	"proofofpeacemaking/internal/core/ports"

	"github.com/gofiber/fiber/v2"
)

type FeedHandler struct {
	feedService ports.FeedService
}

func NewFeedHandler(feedService ports.FeedService) *FeedHandler {
	return &FeedHandler{
		feedService: feedService,
	}
}

func (h *FeedHandler) GetFeed(c *fiber.Ctx) error {
	// Get user address from context (set by auth middleware)
	userAddress := c.Locals("userAddress").(string)

	// Get activities from feed service
	activities, err := h.feedService.GetFeed(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch feed",
		})
	}

	// Render feed template
	return c.Render("feed", fiber.Map{
		"Title":      "Feed - Proof of Peacemaking",
		"Activities": activities,
		"User": map[string]interface{}{
			"Address": userAddress,
		},
	})
}
