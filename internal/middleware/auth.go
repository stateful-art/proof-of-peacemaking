package middleware

import (
	"proofofpeacemaking/internal/core/ports"

	"github.com/gofiber/fiber/v2"
)

type AuthMiddleware struct {
	authService ports.AuthService
}

func NewAuthMiddleware(authService ports.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

func (m *AuthMiddleware) Authenticate() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get token from cookie
		token := c.Cookies("jwt")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Not authenticated",
			})
		}

		// Verify token
		userAddress, err := m.authService.VerifyToken(c.Context(), token)
		if err != nil {
			// Clear invalid cookie
			c.Cookie(&fiber.Cookie{
				Name:     "jwt",
				Value:    "",
				Path:     "/",
				MaxAge:   -1,
				Secure:   true,
				HTTPOnly: true,
				SameSite: "Strict",
			})
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		// Set user address in context
		c.Locals("userAddress", userAddress)
		return c.Next()
	}
}
