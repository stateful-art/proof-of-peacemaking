package handlers

import (
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"
	"time"

	"log"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	// Parse multipart form
	form, err := c.MultipartForm()
	if err != nil {
		log.Printf("[EXPRESSION] Error parsing form: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid form data",
		})
	}

	// Get user from context
	userAddress := c.Locals("userAddress").(string)
	log.Printf("[EXPRESSION] Looking up user with address: %s", userAddress)

	user, err := h.userService.GetUserByAddress(c.Context(), userAddress)
	if err != nil {
		log.Printf("[EXPRESSION] Error getting user: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user",
		})
	}
	if user == nil {
		log.Printf("[EXPRESSION] User not found for address: %s", userAddress)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}
	log.Printf("[EXPRESSION] Found user with ID: %s", user.ID.Hex())

	// Initialize content map
	content := make(map[string]string)

	// Handle text content
	if textContent := form.Value["textContent"]; len(textContent) > 0 {
		content["text"] = textContent[0]
	}

	// Handle image file
	if imageFiles := form.File["imageContent"]; len(imageFiles) > 0 {
		// Save image file
		imageFile := imageFiles[0]
		filename := "uploads/images/" + imageFile.Filename
		if err := c.SaveFile(imageFile, filename); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save image",
			})
		}
		content["image"] = filename
	}

	// Handle audio file
	if audioFiles := form.File["audioContent"]; len(audioFiles) > 0 {
		// Save audio file
		audioFile := audioFiles[0]
		filename := "uploads/audio/" + audioFile.Filename
		if err := c.SaveFile(audioFile, filename); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save audio",
			})
		}
		content["audio"] = filename
	}

	// Handle video file
	if videoFiles := form.File["videoContent"]; len(videoFiles) > 0 {
		// Save video file
		videoFile := videoFiles[0]
		filename := "uploads/video/" + videoFile.Filename
		if err := c.SaveFile(videoFile, filename); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save video",
			})
		}
		content["video"] = filename
	}

	// Create expression domain object
	expression := &domain.Expression{
		ID:             primitive.NewObjectID(),
		Creator:        user.ID.Hex(),
		CreatorAddress: userAddress,
		Content:        content,
		Status:         "pending",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
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
