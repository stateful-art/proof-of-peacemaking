package handlers

import (
	"log"
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	userService  ports.UserService
	statsService ports.StatisticsService
}

func NewUserHandler(userService ports.UserService, statsService ports.StatisticsService) *UserHandler {
	return &UserHandler{
		userService:  userService,
		statsService: statsService,
	}
}

func (h *UserHandler) GetUserService() ports.UserService {
	return h.userService
}

// GetProfile returns the current user's profile
func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	userIdentifier := c.Locals("userAddress").(string)
	var user interface{}
	var err error

	if strings.Contains(userIdentifier, "@") {
		user, err = h.userService.GetUserByEmail(c.Context(), userIdentifier)
	} else {
		user, err = h.userService.GetUserByAddress(c.Context(), userIdentifier)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user profile",
		})
	}

	return c.JSON(user)
}

// UpdateProfile updates the current user's profile
func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	userIdentifier := c.Locals("userAddress").(string)
	var updateData struct {
		Email       string `json:"email"`
		Citizenship string `json:"citizenship"`
		City        string `json:"city"`
	}

	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request data",
		})
	}

	var user *domain.User
	var err error

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

	// Update user fields
	user.Email = updateData.Email
	user.Citizenship = updateData.Citizenship
	user.City = updateData.City

	if err := h.userService.Update(c.Context(), user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update user",
		})
	}

	// Update statistics if citizenship was changed
	if updateData.Citizenship != "" {
		if err := h.statsService.UpdateStatisticsAfterCitizenshipChange(c.Context()); err != nil {
			log.Printf("[USER] Warning: Failed to update statistics: %v", err)
			// Don't return error here, as the user was updated successfully
		}
	}

	return c.JSON(user)
}

// ConnectWallet connects a wallet address to the user's account
func (h *UserHandler) ConnectWallet(c *fiber.Ctx) error {
	var data struct {
		Address   string `json:"address"`
		Signature string `json:"signature"`
		Nonce     int    `json:"nonce"`
	}

	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request data",
		})
	}

	// Implementation depends on your user service methods

	return c.SendStatus(fiber.StatusOK)
}

// GetWalletNonce generates a nonce for wallet connection
func (h *UserHandler) GetWalletNonce(c *fiber.Ctx) error {
	var data struct {
		Address string `json:"address"`
	}

	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request data",
		})
	}

	// Implementation depends on your user service methods

	return c.JSON(fiber.Map{
		"nonce": 123, // Replace with actual nonce generation
	})
}
