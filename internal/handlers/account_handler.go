package handlers

import (
	"log"
	"proofofpeacemaking/internal/core/ports"

	"github.com/gofiber/fiber/v2"
)

type AccountHandler struct {
	userService ports.UserService
}

func NewAccountHandler(userService ports.UserService) *AccountHandler {
	return &AccountHandler{
		userService: userService,
	}
}

func (h *AccountHandler) HandleAccount(c *fiber.Ctx) error {
	userAddress := c.Locals("userAddress").(string)
	if userAddress == "" {
		return c.Redirect("/")
	}

	user, err := h.userService.GetUserByAddress(c.Context(), userAddress)
	if err != nil {
		log.Printf("Error getting user: %v", err)
		return c.Render("error", fiber.Map{
			"Error": "Failed to get user data",
		})
	}

	return c.Render("account", fiber.Map{
		"Title": "Account Settings",
		"User":  user,
	})
}

func (h *AccountHandler) UpdateProfile(c *fiber.Ctx) error {
	userAddress := c.Locals("userAddress").(string)
	if userAddress == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Get current user
	user, err := h.userService.GetUserByAddress(c.Context(), userAddress)
	if err != nil {
		log.Printf("Error getting user: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user data",
		})
	}

	// Parse update data
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

	// Update user fields
	user.Email = updateData.Email
	user.Citizenship = updateData.Citizenship
	user.City = updateData.City

	// Validate and update user
	if err := h.userService.Update(c.Context(), user); err != nil {
		if err.Error() == "email already exists" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"errors": map[string]string{
					"email": "Email is already in use",
				},
			})
		}
		log.Printf("Error updating user: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update profile",
		})
	}

	return c.JSON(user)
}

func (h *AccountHandler) ConnectWallet(c *fiber.Ctx) error {
	var data struct {
		Address string `json:"address"`
	}

	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request data",
		})
	}

	userAddress := c.Locals("userAddress").(string)
	user, err := h.userService.GetUserByAddress(c.Context(), userAddress)
	if err != nil {
		log.Printf("Error getting user: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user data",
		})
	}

	if err := h.userService.ConnectWallet(c.Context(), user.ID, data.Address); err != nil {
		if err.Error() == "wallet already connected to another account" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Wallet is already connected to another account",
			})
		}
		log.Printf("Error connecting wallet: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to connect wallet",
		})
	}

	return c.SendStatus(fiber.StatusOK)
}
