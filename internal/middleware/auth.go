package middleware

import (
	"log"
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
		token := c.Cookies("session")
		if token == "" {
			log.Printf("\n\n\n\n\n\n[AUTH] No token found in cookie\n\n\n\n\n\n")
			// For API routes, return JSON error
			if c.Path() == "/api" {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Not authenticated",
				})
			}
			// For page routes, redirect to home
			return c.Redirect("/")
		}

		// Verify token
		userAddress, err := m.authService.VerifyToken(c.Context(), token)
		if err != nil {
			// Clear invalid cookie
			c.Cookie(&fiber.Cookie{
				Name:     "session",
				Value:    "",
				Path:     "/",
				MaxAge:   -1,
				Secure:   true,
				HTTPOnly: true,
				SameSite: "Strict",
			})
			// For API routes, return JSON error
			if c.Path() == "/api" {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Invalid or expired token",
				})
			}
			log.Printf("\n\n\n\n\n\n[AUTH] Redirecting to home\n\n\n\n\n\n")
			// For page routes, redirect to home
			return c.Redirect("/")
		}

		// Set user address in context
		c.Locals("userAddress", userAddress)
		log.Printf("\n\n\n\n\n\n[AUTH] User address set in context: %s\n\n\n\n\n\n", userAddress)
		return c.Next()
	}
}
