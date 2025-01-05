package handlers

import (
	"log"
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type AccountHandler struct {
	userService  ports.UserService
	authService  ports.AuthService
	statsService ports.StatisticsService
}

func NewAccountHandler(userService ports.UserService, authService ports.AuthService, statsService ports.StatisticsService) *AccountHandler {
	return &AccountHandler{
		userService:  userService,
		authService:  authService,
		statsService: statsService,
	}
}

func (h *AccountHandler) HandleAccount(c *fiber.Ctx) error {
	userIdentifier := c.Locals("userAddress").(string)
	if userIdentifier == "" {
		return c.Redirect("/")
	}

	var user *domain.User
	var err error

	// Try to get user by wallet address first
	user, err = h.userService.GetUserByAddress(c.Context(), userIdentifier)
	if err != nil || user == nil {
		// If not found by address, try by email
		user, err = h.userService.GetUserByEmail(c.Context(), userIdentifier)
		if err != nil {
			log.Printf("Error getting user: %v", err)
			return c.Render("error", fiber.Map{
				"Error": "Failed to get user data",
			})
		}
	}

	if user == nil {
		log.Printf("User not found for identifier: %s", userIdentifier)
		return c.Render("error", fiber.Map{
			"Error": "User not found",
		})
	}

	return c.Render("account", fiber.Map{
		"Title": "Account Settings",
		"User":  user,
	})
}

func (h *AccountHandler) UpdateProfile(c *fiber.Ctx) error {
	userIdentifier := c.Locals("userAddress").(string)

	// Try to get user by wallet address first
	user, err := h.userService.GetUserByAddress(c.Context(), userIdentifier)
	if err != nil || user == nil {
		// If not found by address, try by email
		user, err = h.userService.GetUserByEmail(c.Context(), userIdentifier)
		if err != nil {
			log.Printf("Error getting user: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get user data",
			})
		}
	}

	if user == nil {
		log.Printf("User not found for identifier: %s", userIdentifier)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

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

	// Store old citizenship to check if it changed
	oldCitizenship := user.Citizenship

	// Update user fields
	user.Email = updateData.Email
	user.Citizenship = updateData.Citizenship
	user.City = updateData.City

	// Validate and update user
	if err := h.userService.Update(c.Context(), user); err != nil {
		log.Printf("Error updating user: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update user",
		})
	}

	// If citizenship changed, update statistics
	if oldCitizenship != user.Citizenship {
		if err := h.statsService.UpdateStatisticsAfterCitizenshipChange(c.Context()); err != nil {
			log.Printf("Error updating statistics after citizenship change: %v", err)
			// Don't return error here as the user update was successful
		}
	}

	return c.JSON(fiber.Map{
		"message": "Profile updated successfully",
	})
}

func (h *AccountHandler) GetWalletNonce(c *fiber.Ctx) error {
	var data struct {
		Address string `json:"address"`
	}

	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request data",
		})
	}

	// Check if wallet is already connected to another user
	existingUser, err := h.userService.GetUserByAddress(c.Context(), data.Address)
	if err != nil {
		log.Printf("Error checking existing wallet: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check wallet status",
		})
	}

	if existingUser != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Wallet is already connected to another account",
		})
	}

	// Generate nonce for the wallet
	nonce, err := h.authService.GenerateNonce(c.Context(), data.Address)
	if err != nil {
		log.Printf("Error generating nonce: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate nonce",
		})
	}

	return c.JSON(fiber.Map{
		"nonce": nonce,
	})
}

func (h *AccountHandler) ConnectWallet(c *fiber.Ctx) error {
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

	userIdentifier := c.Locals("userAddress").(string)

	// Try to get user by wallet address first
	user, err := h.userService.GetUserByAddress(c.Context(), userIdentifier)
	if err != nil || user == nil {
		// If not found by address, try by email
		user, err = h.userService.GetUserByEmail(c.Context(), userIdentifier)
		if err != nil {
			log.Printf("Error getting user: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get user data",
			})
		}
	}

	if user == nil {
		log.Printf("User not found for identifier: %s", userIdentifier)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Verify signature
	verified, storedNonce, err := h.authService.VerifySignature(c.Context(), data.Address, data.Signature)
	if err != nil {
		log.Printf("Error verifying signature: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to verify wallet ownership",
		})
	}

	if !verified {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid signature",
		})
	}

	// Convert stored nonce to int
	storedNonceInt, err := strconv.Atoi(storedNonce)
	if err != nil {
		log.Printf("Error converting nonce: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to verify nonce",
		})
	}

	// Verify nonce
	if storedNonceInt != data.Nonce {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid nonce",
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
