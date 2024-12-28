package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/stateful-art/proof-of-peacemaking/internal/core/ports"
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
	address := c.Query("address")
	if address == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Address is required",
		})
	}

	nonce, err := h.authService.GenerateNonce(c.Context(), address)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate nonce",
		})
	}

	return c.JSON(fiber.Map{
		"nonce": nonce,
	})
}

func (h *AuthHandler) VerifySignature(c *fiber.Ctx) error {
	var body struct {
		Address   string `json:"address"`
		Signature string `json:"signature"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	isValid, err := h.authService.VerifySignature(c.Context(), body.Address, body.Signature)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to verify signature",
		})
	}

	return c.JSON(fiber.Map{
		"valid": isValid,
	})
}
