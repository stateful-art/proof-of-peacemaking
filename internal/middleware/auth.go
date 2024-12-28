package middleware

import (
	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get token from header
		token := c.Get("Authorization")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing authorization token",
			})
		}

		// Verify token and set user in context
		// TODO: Implement actual token verification
		userAddress := "0x..." // Get from token
		c.Locals("userAddress", userAddress)

		return c.Next()
	}
}
