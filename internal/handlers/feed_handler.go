package handlers

import (
	"log"
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
	log.Printf("[FEED] Starting feed handler")

	// Get user address from context (set by auth middleware)
	userAddress, ok := c.Locals("userAddress").(string)
	if !ok {
		log.Printf("[FEED] Error: User address not found in context")
		// Redirect to home page if not authenticated
		return c.Redirect("/")
	}
	log.Printf("[FEED] Got user address: %s", userAddress)

	// Get activities from feed service
	activities, err := h.feedService.GetFeed(c.Context())
	if err != nil {
		log.Printf("[FEED] Error fetching feed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch feed",
		})
	}

	log.Printf("[FEED] Rendering feed for user %s with %d activities", userAddress, len(activities))
	data := fiber.Map{
		"Title":      "Feed - Proof of Peacemaking",
		"Activities": activities,
		"User": map[string]interface{}{
			"Address": userAddress,
		},
	}
	log.Printf("[FEED] Data being passed to template: %+v", data)

	return c.Render("feed", data, "")
}
