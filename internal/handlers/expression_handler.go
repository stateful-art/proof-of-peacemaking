package handlers

import (
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"
	"time"

	"log"

	"strings"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Helper function to get map keys
func getKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

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

	// Log form contents (only field names and sizes)
	log.Printf("[EXPRESSION] Form fields: %v", getKeys(form.Value))
	log.Printf("[EXPRESSION] File fields: %v", getKeys(form.File))
	log.Printf("[EXPRESSION] Number of image files: %d", len(form.File["imageContent"]))
	log.Printf("[EXPRESSION] Number of audio files: %d", len(form.File["audioContent"]))
	log.Printf("[EXPRESSION] Number of video files: %d", len(form.File["videoContent"]))
	for key, files := range form.File {
		for _, file := range files {
			log.Printf("[EXPRESSION] File %s: name=%s, size=%d", key, file.Filename, file.Size)
		}
	}

	// Get user from context
	userIdentifier := c.Locals("userAddress").(string)
	log.Printf("[EXPRESSION] Looking up user with identifier: %s", userIdentifier)

	// Get user by email or address
	var user *domain.User
	if strings.Contains(userIdentifier, "@") {
		user, err = h.userService.GetUserByEmail(c.Context(), userIdentifier)
	} else {
		user, err = h.userService.GetUserByAddress(c.Context(), userIdentifier)
	}

	if err != nil {
		log.Printf("[EXPRESSION] Error getting user: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user",
		})
	}
	if user == nil {
		log.Printf("[EXPRESSION] User not found for identifier: %s", userIdentifier)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}
	log.Printf("[EXPRESSION] Found user with ID: %s", user.ID.Hex())

	// Initialize content map
	content := make(map[string]string)

	// Create expression domain object first to get the ID
	expression := &domain.Expression{
		ID:             primitive.NewObjectID(),
		Creator:        user.ID.Hex(),
		CreatorAddress: userIdentifier,
		Content:        content,
		Status:         "pending",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Handle text content
	if textContent := form.Value["textContent"]; len(textContent) > 0 {
		content["text"] = textContent[0]
	}

	// Handle image file
	if imageFiles := form.File["imageContent"]; len(imageFiles) > 0 {
		imageFile := imageFiles[0]
		file, err := imageFile.Open()
		if err != nil {
			log.Printf("[EXPRESSION] Error opening image file: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to process image",
			})
		}
		defer file.Close()

		// Upload to R2
		key, err := h.expressionService.UploadMedia(c.Context(), expression.ID.Hex(), "image", file, imageFile.Filename)
		if err != nil {
			log.Printf("[EXPRESSION] Error uploading image to R2: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to upload image",
			})
		}
		content["image"] = key
	}

	// Handle audio file
	if audioFiles := form.File["audioContent"]; len(audioFiles) > 0 {
		audioFile := audioFiles[0]
		file, err := audioFile.Open()
		if err != nil {
			log.Printf("[EXPRESSION] Error opening audio file: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to process audio",
			})
		}
		defer file.Close()

		// Upload to R2
		key, err := h.expressionService.UploadMedia(c.Context(), expression.ID.Hex(), "audio", file, audioFile.Filename)
		if err != nil {
			log.Printf("[EXPRESSION] Error uploading audio to R2: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to upload audio",
			})
		}
		content["audio"] = key
	}

	// Handle video file
	if videoFiles := form.File["videoContent"]; len(videoFiles) > 0 {
		videoFile := videoFiles[0]
		file, err := videoFile.Open()
		if err != nil {
			log.Printf("[EXPRESSION] Error opening video file: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to process video",
			})
		}
		defer file.Close()

		// Upload to R2
		key, err := h.expressionService.UploadMedia(c.Context(), expression.ID.Hex(), "video", file, videoFile.Filename)
		if err != nil {
			log.Printf("[EXPRESSION] Error uploading video to R2: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to upload video",
			})
		}
		content["video"] = key
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
