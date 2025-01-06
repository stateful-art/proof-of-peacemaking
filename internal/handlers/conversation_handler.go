package handlers

import (
	"log"
	"os"
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ConversationHandler struct {
	service     ports.ConversationService
	userService ports.UserService
}

func NewConversationHandler(service ports.ConversationService, userService ports.UserService) *ConversationHandler {
	return &ConversationHandler{
		service:     service,
		userService: userService,
	}
}

type CreateConversationRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ImageURL    string    `json:"imageUrl"`
	StartTime   time.Time `json:"startTime"`
	Tags        []string  `json:"tags"`
}

func (h *ConversationHandler) CreateConversation(c *fiber.Ctx) error {
	var req CreateConversationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	userIdentifier := c.Locals("userAddress").(string)
	var user *domain.User
	var err error

	// Get user by email or address
	if strings.Contains(userIdentifier, "@") {
		user, err = h.userService.GetUserByEmail(c.Context(), userIdentifier)
	} else {
		user, err = h.userService.GetUserByAddress(c.Context(), userIdentifier)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user",
		})
	}

	if user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	conversation := &domain.Conversation{
		Title:       req.Title,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		CreatorID:   user.ID.Hex(), // Use user ID instead of email
		StartTime:   req.StartTime,
		Tags:        req.Tags,
	}

	if err := h.service.CreateConversation(c.Context(), conversation); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(conversation)
}

func (h *ConversationHandler) ListConversations(c *fiber.Ctx) error {
	filter := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filter["status"] = status
	}

	conversations, err := h.service.ListConversations(c.Context(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(conversations)
}

func (h *ConversationHandler) GetConversation(c *fiber.Ctx) error {
	id, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID",
		})
	}

	conversation, err := h.service.GetConversation(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if conversation == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Conversation not found",
		})
	}

	return c.JSON(conversation)
}

func (h *ConversationHandler) StartConversation(c *fiber.Ctx) error {
	id, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		log.Printf("[ERROR] Invalid conversation ID: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID",
		})
	}

	userID := c.Locals("userAddress").(string)
	log.Printf("[DEBUG] Starting conversation. ID: %s, User: %s", id.Hex(), userID)

	// Get user by email or address
	var user *domain.User
	if strings.Contains(userID, "@") {
		user, err = h.userService.GetUserByEmail(c.Context(), userID)
	} else {
		user, err = h.userService.GetUserByAddress(c.Context(), userID)
	}

	if err != nil {
		log.Printf("[ERROR] Failed to get user: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user",
		})
	}

	if user == nil {
		log.Printf("[ERROR] User not found: %s", userID)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Use user ID for comparison
	if err := h.service.StartConversation(c.Context(), id, user.ID.Hex()); err != nil {
		log.Printf("[ERROR] Failed to start conversation: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	log.Printf("[INFO] Successfully started conversation %s by user %s", id.Hex(), user.ID.Hex())
	return c.SendStatus(fiber.StatusOK)
}

func (h *ConversationHandler) EndConversation(c *fiber.Ctx) error {
	id, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID",
		})
	}

	userID := c.Locals("userAddress").(string)
	if err := h.service.EndConversation(c.Context(), id, userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusOK)
}

func (h *ConversationHandler) SubscribeToNotifications(c *fiber.Ctx) error {
	id, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID",
		})
	}

	userID := c.Locals("userAddress").(string)
	if err := h.service.SubscribeToNotifications(c.Context(), id, userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusOK)
}

func (h *ConversationHandler) UnsubscribeFromNotifications(c *fiber.Ctx) error {
	id, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID",
		})
	}

	userID := c.Locals("userAddress").(string)
	if err := h.service.UnsubscribeFromNotifications(c.Context(), id, userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusOK)
}

func (h *ConversationHandler) GenerateJoinToken(c *fiber.Ctx) error {
	id, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID",
		})
	}

	userID := c.Locals("userAddress").(string)
	conversation, err := h.service.GetConversation(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if conversation == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Conversation not found",
		})
	}

	// Only creator can publish
	canPublish := conversation.CreatorID == userID

	token, err := h.service.GenerateJoinToken(userID, conversation.RoomName, canPublish)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"token": token,
	})
}

func (h *ConversationHandler) ServeRoom(c *fiber.Ctx) error {
	userIdentifier := c.Locals("userAddress").(string)
	conversationID := c.Params("id")

	// Get user by email or address
	var user *domain.User
	var err error
	if strings.Contains(userIdentifier, "@") {
		user, err = h.userService.GetUserByEmail(c.Context(), userIdentifier)
	} else {
		user, err = h.userService.GetUserByAddress(c.Context(), userIdentifier)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user data",
		})
	}

	if user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Convert string ID to ObjectID
	id, err := primitive.ObjectIDFromHex(conversationID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid conversation ID",
		})
	}

	// Get conversation details
	conversation, err := h.service.GetConversation(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get conversation details",
		})
	}

	if conversation == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Conversation not found",
		})
	}

	// Check if conversation is live
	if conversation.Status != domain.ConversationStatusLive {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Conversation is not live",
		})
	}

	// Get join token
	token, err := h.service.GenerateJoinToken(user.ID.Hex(), conversation.RoomName, conversation.CreatorID == user.ID.Hex())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate join token",
		})
	}

	// Prepare data for the template
	data := fiber.Map{
		"Title":        conversation.Title,
		"Conversation": conversation,
		"Token":        token,
		"RoomName":     conversation.RoomName,
		"IsCreator":    conversation.CreatorID == user.ID.Hex(),
		"User": fiber.Map{
			"ID":      user.ID.Hex(),
			"Email":   user.Email,
			"Address": user.Address,
		},
		"LiveKitHost": os.Getenv("LIVEKIT_HOST"),
	}

	return c.Render("conversation_room", data)
}
