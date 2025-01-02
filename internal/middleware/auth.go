package middleware

import (
	"log"
	"proofofpeacemaking/internal/core/ports"
	"strings"

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
			log.Printf("[AUTH] No token found in cookie")
			// For API routes, return JSON error
			if strings.HasPrefix(c.Path(), "/api") {
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
			log.Printf("[AUTH] Token verification failed: %v", err)
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
			if strings.HasPrefix(c.Path(), "/api") {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Invalid or expired token",
				})
			}
			// For page routes, redirect to home
			return c.Redirect("/")
		}

		// Set user address in context
		c.Locals("userAddress", userAddress)
		log.Printf("[AUTH] User address set in context: %s", userAddress)
		return c.Next()
	}
}

func (m *AuthMiddleware) Optional() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get session cookie
		sessionCookie := c.Cookies("session")
		if sessionCookie == "" {
			// No session cookie, continue without user data
			return c.Next()
		}

		// Verify session token
		userIdentifier, err := m.authService.VerifyToken(c.Context(), sessionCookie)
		if err != nil {
			// Invalid session, continue without user data
			return c.Next()
		}

		// Set user identifier in context
		c.Locals("userAddress", userIdentifier)
		return c.Next()
	}
}
