package handlers

import (
	"log"
	"proofofpeacemaking/internal/core/ports"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService ports.AuthService
}

func NewAuthHandler(authService ports.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) GenerateNonce(c *fiber.Ctx) error {
	log.Printf("[AUTH] Generating nonce - Request Method: %s", c.Method())
	address := c.Query("address")
	log.Printf("[AUTH] Address from query: %s", address)

	if address == "" {
		log.Printf("[AUTH] Error: Missing address in request")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Address is required",
		})
	}

	nonce, err := h.authService.GenerateNonce(c.Context(), address)
	if err != nil {
		log.Printf("[AUTH] Error generating nonce: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	log.Printf("[AUTH] Successfully generated nonce %d for address %s", nonce, address)
	return c.JSON(fiber.Map{
		"nonce": nonce,
	})
}

func (h *AuthHandler) VerifySignature(c *fiber.Ctx) error {
	log.Printf("[AUTH] Verifying signature - Request Method: %s", c.Method())
	var body struct {
		Address   string `json:"address"`
		Signature string `json:"signature"`
	}

	if err := c.BodyParser(&body); err != nil {
		log.Printf("[AUTH] Error parsing request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	log.Printf("[AUTH] Verifying signature for address: %s", body.Address)
	isValid, token, err := h.authService.VerifySignature(c.Context(), body.Address, body.Signature)
	if err != nil {
		log.Printf("[AUTH] Error verifying signature: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if isValid && token != "" {
		log.Printf("[AUTH] Signature verified successfully for address: %s", body.Address)
		// Set secure HTTP-only cookie
		cookie := fiber.Cookie{
			Name:     "jwt",
			Value:    token,
			Path:     "/",
			MaxAge:   24 * 60 * 60, // 24 hours
			Secure:   true,         // Only send over HTTPS
			HTTPOnly: true,         // Prevent JavaScript access
			SameSite: "Strict",     // CSRF protection
		}
		c.Cookie(&cookie)
	} else {
		log.Printf("[AUTH] Invalid signature for address: %s", body.Address)
	}

	return c.JSON(fiber.Map{
		"valid": isValid,
	})
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var body struct {
		Address string `json:"address"`
		Email   string `json:"email"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if body.Address == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Address is required",
		})
	}

	user, token, err := h.authService.Register(c.Context(), body.Address, body.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to register user",
		})
	}

	// Set secure HTTP-only cookie
	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    token,
		Path:     "/",
		MaxAge:   24 * 60 * 60, // 24 hours
		Secure:   true,         // Only send over HTTPS
		HTTPOnly: true,         // Prevent JavaScript access
		SameSite: "Strict",     // CSRF protection
	}
	c.Cookie(&cookie)

	return c.JSON(fiber.Map{
		"user": user,
	})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// Clear the JWT cookie
	c.Cookie(&fiber.Cookie{
		Name:     "jwt",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   true,
		HTTPOnly: true,
		SameSite: "Strict",
	})

	return c.SendStatus(fiber.StatusOK)
}

func (h *AuthHandler) GetAuthService() ports.AuthService {
	return h.authService
}
