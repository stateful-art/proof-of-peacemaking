package handlers

import (
	"log"
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"

	"github.com/gofiber/fiber/v2"
)

type FeedHandler struct {
	feedService ports.FeedService
	userService ports.UserService
}

func NewFeedHandler(feedService ports.FeedService, userService ports.UserService) *FeedHandler {
	return &FeedHandler{
		feedService: feedService,
		userService: userService,
	}
}

func (h *FeedHandler) GetFeedService() ports.FeedService {
	return h.feedService
}

func (h *FeedHandler) HandleFeed(c *fiber.Ctx) error {
	userAddress := c.Locals("userAddress").(string)
	if userAddress == "" {
		return c.Redirect("/")
	}

	// Get user ID for acknowledgement comparison
	user, err := h.userService.GetUserByAddress(c.Context(), userAddress)
	if err != nil {
		log.Printf("Error getting user: %v", err)
		return c.Render("error", fiber.Map{
			"Error": "Failed to get user data",
		})
	}

	activities, err := h.feedService.GetFeed(c.Context())
	if err != nil {
		log.Printf("Error getting feed: %v", err)
		return c.Render("error", fiber.Map{
			"Error": "Failed to load feed",
		})
	}

	// Map the activities to include CreatorAddress and acknowledgement status
	expressions := make([]fiber.Map, len(activities))
	for i, activity := range activities {
		// Check if the current user has acknowledged this expression
		acks := activity["Acknowledgements"].([]*domain.Acknowledgement)
		isAcknowledged := false
		activeCount := 0
		for _, ack := range acks {
			if ack.Status == domain.AcknowledgementStatusActive {
				activeCount++
				if ack.Acknowledger == user.ID.Hex() {
					isAcknowledged = true
				}
			}
		}

		expressions[i] = fiber.Map{
			"ID":                         activity["ID"],
			"CreatorAddress":             activity["CreatorAddress"],
			"Content":                    activity["Content"],
			"Timestamp":                  activity["Timestamp"],
			"IsAcknowledged":             isAcknowledged,
			"ActiveAcknowledgementCount": activeCount,
		}
	}

	data := fiber.Map{
		"Title":       "Feed",
		"UserAddress": userAddress,
		"Expressions": expressions,
	}

	log.Printf("[FEED] Data being passed to template: %+v", data)
	return c.Render("feed", data)
}
