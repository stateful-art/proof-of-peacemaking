package handlers

import (
	"log"
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"
	"strings"
	"time"

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
	userIdentifier := c.Locals("userAddress").(string)
	if userIdentifier == "" {
		return c.Redirect("/")
	}

	// Get user ID for acknowledgement comparison
	var user *domain.User
	var err error

	// Check if the identifier is an email or wallet address
	if strings.Contains(userIdentifier, "@") {
		user, err = h.userService.GetUserByEmail(c.Context(), userIdentifier)
	} else {
		user, err = h.userService.GetUserByAddress(c.Context(), userIdentifier)
	}

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
		hasActiveAck := false
		activeCount := 0
		userAckStatus := ""

		for _, ack := range acks {
			if ack.Status == domain.AcknowledgementStatusActive {
				activeCount++
			}
			// Track the user's acknowledgment status regardless of whether it's active
			if ack.Acknowledger == user.ID.Hex() {
				userAckStatus = string(ack.Status)
				if ack.Status == domain.AcknowledgementStatusActive {
					hasActiveAck = true
				}
			}
		}

		// Format timestamp
		timestamp := activity["Timestamp"].(time.Time)
		formattedTime := timestamp.Format("Jan 02, 2006 15:04")

		expressions[i] = fiber.Map{
			"ID":                         activity["ID"],
			"CreatorAddress":             activity["CreatorAddress"],
			"Content":                    activity["Content"],
			"Timestamp":                  formattedTime,
			"HasActiveAck":               hasActiveAck,
			"UserAckStatus":              userAckStatus,
			"ActiveAcknowledgementCount": activeCount,
		}
	}

	data := fiber.Map{
		"Title":       "Feed",
		"User":        fiber.Map{"Email": user.Email, "Address": user.Address},
		"Expressions": expressions,
	}

	log.Printf("[FEED] Data being passed to template: %+v", data)
	return c.Render("feed", data)
}
