package handlers

import (
	"log"
	"proofofpeacemaking/internal/core/ports"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService ports.AuthService
	userService ports.UserService
}

func NewAuthHandler(authService ports.AuthService, userService ports.UserService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
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
	var body struct {
		Address   string `json:"address"`
		Signature string `json:"signature"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	isValid, token, err := h.authService.VerifySignature(c.Context(), body.Address, body.Signature)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if !isValid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid signature",
		})
	}

	// Set session cookie
	cookie := fiber.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		MaxAge:   24 * 60 * 60, // 24 hours
		Secure:   true,
		HTTPOnly: true,
		SameSite: "Strict",
	}
	c.Cookie(&cookie)

	return c.JSON(fiber.Map{
		"valid":    true,
		"token":    token,
		"redirect": "/feed", // Add redirect URL
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
		Name:     "session",
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
	// Get session token from cookie
	sessionToken := c.Cookies("session")
	if sessionToken != "" {
		// Get user identifier before invalidating session
		userIdentifier := ""
		if userAddr := c.Locals("userAddress"); userAddr != nil {
			userIdentifier = userAddr.(string)
		}

		// Invalidate session in database
		if err := h.authService.Logout(c.Context(), sessionToken); err != nil {
			log.Printf("[AUTH] Error invalidating session: %v", err)
			// Continue with cookie cleanup even if session invalidation fails
		}

		// Only try to delete all sessions if we have a user identifier
		if userIdentifier != "" {
			if err := h.authService.DeleteAllUserSessions(c.Context(), userIdentifier); err != nil {
				log.Printf("[AUTH] Error deleting user sessions: %v", err)
				// Continue as this is not critical
			}
		}
	}

	// Clear all auth-related cookies with all possible paths
	cookiesToClear := []string{"session", "user", "token"}
	paths := []string{"/", "/auth", "/api"}

	for _, cookieName := range cookiesToClear {
		for _, path := range paths {
			c.Cookie(&fiber.Cookie{
				Name:     cookieName,
				Value:    "",
				Path:     path,
				MaxAge:   -1,
				Expires:  time.Now().Add(-24 * time.Hour), // Set expiry in the past
				Secure:   true,
				HTTPOnly: true,
				SameSite: "Strict",
			})
		}
	}

	return c.SendStatus(fiber.StatusOK)
}

func (h *AuthHandler) GetSession(c *fiber.Ctx) error {
	// Get session token from cookie
	sessionCookie := c.Cookies("session")
	if sessionCookie == "" {
		log.Printf("[AUTH] No session cookie found")
		return c.JSON(fiber.Map{
			"authenticated": false,
		})
	}

	// Verify session token and get address
	address, err := h.authService.VerifyToken(c.Context(), sessionCookie)
	if err != nil {
		log.Printf("[AUTH] Session verification failed: %v", err)
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
		return c.JSON(fiber.Map{
			"authenticated": false,
			"error":         "Invalid session",
		})
	}

	// log.Printf("[AUTH] Session verified for address: %s", address)
	return c.JSON(fiber.Map{
		"authenticated": true,
		"address":       address,
	})
}

func (h *AuthHandler) GetAuthService() ports.AuthService {
	return h.authService
}

func (h *AuthHandler) RegisterWithEmail(c *fiber.Ctx) error {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Username string `json:"username"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if body.Email == "" || body.Password == "" || body.Username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email, password and username are required",
		})
	}

	user, token, err := h.authService.RegisterWithEmail(c.Context(), body.Email, body.Password, body.Username)
	if err != nil {
		// Check if the error is a duplicate key error
		if strings.Contains(err.Error(), "E11000 duplicate key error") {
			// Check if the user was actually created (race condition where one request succeeded)
			existingUser, checkErr := h.userService.GetUserByEmail(c.Context(), body.Email)
			if checkErr == nil && existingUser != nil {
				// User exists and was created by the parallel request
				// Generate a new session token for this user
				loginUser, newToken, loginErr := h.authService.LoginWithEmail(c.Context(), body.Email, body.Password)
				if loginErr == nil {
					// Set the cookie and return success
					cookie := fiber.Cookie{
						Name:     "session",
						Value:    newToken,
						Path:     "/",
						MaxAge:   24 * 60 * 60, // 24 hours
						Secure:   true,         // Only send over HTTPS
						HTTPOnly: true,         // Prevent JavaScript access
						SameSite: "Strict",     // CSRF protection
					}
					c.Cookie(&cookie)

					return c.JSON(fiber.Map{
						"user":     loginUser,
						"redirect": "/feed",
						"message":  "Registration successful! Redirecting in 10 seconds...",
					})
				}
			}

			// If we couldn't verify the user exists, return the appropriate error
			if strings.Contains(err.Error(), "email_1") {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"error": "Email already registered",
				})
			} else if strings.Contains(err.Error(), "username_1") {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"error": "Username already taken",
				})
			}
		}

		log.Printf("[AUTH] Registration error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Set secure HTTP-only cookie
	cookie := fiber.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		MaxAge:   24 * 60 * 60, // 24 hours
		Secure:   true,         // Only send over HTTPS
		HTTPOnly: true,         // Prevent JavaScript access
		SameSite: "Strict",     // CSRF protection
	}
	c.Cookie(&cookie)

	return c.JSON(fiber.Map{
		"user":     user,
		"redirect": "/feed",
		"message":  "Registration successful! Redirecting in 10 seconds...",
	})
}

func (h *AuthHandler) LoginWithEmail(c *fiber.Ctx) error {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if body.Email == "" || body.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email and password are required",
		})
	}

	user, token, err := h.authService.LoginWithEmail(c.Context(), body.Email, body.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid email or password",
		})
	}

	// Set secure HTTP-only cookie
	cookie := fiber.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		MaxAge:   24 * 60 * 60, // 24 hours
		Secure:   true,         // Only send over HTTPS
		HTTPOnly: true,         // Prevent JavaScript access
		SameSite: "Strict",     // CSRF protection
	}
	c.Cookie(&cookie)

	return c.JSON(fiber.Map{
		"user":     user,
		"redirect": "/feed",
	})
}
