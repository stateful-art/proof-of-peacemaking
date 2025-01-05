package handlers

import (
	"bytes"
	"encoding/json"
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"
	"strings"

	"time"

	"log"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WebAuthnHandler struct {
	webAuthnService ports.WebAuthnService
	sessionService  ports.SessionService
	userService     ports.UserService
}

func NewWebAuthnHandler(webAuthnService ports.WebAuthnService, sessionService ports.SessionService, userService ports.UserService) *WebAuthnHandler {
	return &WebAuthnHandler{
		webAuthnService: webAuthnService,
		sessionService:  sessionService,
		userService:     userService,
	}
}

// BeginRegistration initiates the passkey registration process
func (h *WebAuthnHandler) BeginRegistration(c *fiber.Ctx) error {
	// Parse registration request
	var req struct {
		Email    string `json:"email"`
		Username string `json:"username"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Check if email or username already exists
	existingUser, err := h.userService.GetUserByEmail(c.Context(), req.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to check email",
		})
	}
	if existingUser != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "email already registered",
		})
	}

	existingUser, err = h.userService.GetUserByUsername(c.Context(), req.Username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to check username",
		})
	}
	if existingUser != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "username already taken",
		})
	}

	// Create a new user
	user := &domain.User{
		ID:        primitive.NewObjectID(),
		Email:     req.Email,
		Username:  req.Username,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save user to database
	if err := h.userService.Create(c.Context(), user); err != nil {
		if strings.HasPrefix(err.Error(), "validation failed:") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		log.Printf("Failed to create user: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create user",
		})
	}

	// Begin registration
	options, sessionData, err := h.webAuthnService.BeginRegistration(c.Context(), user.ID)
	if err != nil {
		// Log the error but continue - user will be cleaned up by a background job if registration isn't completed
		log.Printf("Failed to begin WebAuthn registration for user %s: %v", user.ID.Hex(), err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Create a temporary session for storing WebAuthn data
	session := &domain.Session{
		UserID:       user.ID.Hex(),
		WebAuthnData: string(must(json.Marshal(sessionData))),
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(5 * time.Minute), // Short expiry for registration
	}

	if err := h.sessionService.Create(c.Context(), session); err != nil {
		// Log the error but continue - user will be cleaned up by a background job if registration isn't completed
		log.Printf("Failed to create session for user %s: %v", user.ID.Hex(), err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create session",
		})
	}

	// Set session cookie
	c.Cookie(&fiber.Cookie{
		Name:     "session",
		Value:    session.Token,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
	})

	// Add user data to the response
	options.Response.User.DisplayName = user.Username
	options.Response.User.Name = user.Email

	return c.JSON(options)
}

// Helper function for JSON marshaling
func must(data []byte, err error) []byte {
	if err != nil {
		panic(err)
	}
	return data
}

// FinishRegistration completes the passkey registration process
func (h *WebAuthnHandler) FinishRegistration(c *fiber.Ctx) error {
	// Get user from session
	session, err := h.sessionService.GetSession(c.Context(), c.Cookies("session"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	userID, err := primitive.ObjectIDFromHex(session.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user ID",
		})
	}

	// Get session data from user's session
	if session.WebAuthnData == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "no session data found",
		})
	}

	var sessionData webauthn.SessionData
	if err := json.Unmarshal([]byte(session.WebAuthnData), &sessionData); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to deserialize session data",
		})
	}

	// Parse response
	response, err := protocol.ParseCredentialCreationResponseBody(bytes.NewReader(c.Body()))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "failed to parse response",
		})
	}

	// Complete registration
	if err := h.webAuthnService.FinishRegistration(c.Context(), userID, sessionData, response); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Clear session data
	session.WebAuthnData = ""
	if err := h.sessionService.Update(c.Context(), session); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update session",
		})
	}

	return c.JSON(fiber.Map{
		"message": "passkey registered successfully",
	})
}

// BeginAuthentication initiates the passkey authentication process
func (h *WebAuthnHandler) BeginAuthentication(c *fiber.Ctx) error {
	// Get user ID from request
	userID, err := primitive.ObjectIDFromHex(c.Query("user_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user ID",
		})
	}

	// Begin authentication
	options, sessionData, err := h.webAuthnService.BeginAuthentication(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Create a temporary session for storing WebAuthn session data
	sessionDataJSON, err := json.Marshal(sessionData)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to serialize session data",
		})
	}

	session := &domain.Session{
		UserID:       userID.Hex(),
		WebAuthnData: string(sessionDataJSON),
	}

	if err := h.sessionService.Create(c.Context(), session); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to store session data",
		})
	}

	// Set session cookie
	c.Cookie(&fiber.Cookie{
		Name:     "session",
		Value:    session.Token,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
	})

	return c.JSON(options)
}

// FinishAuthentication completes the passkey authentication process
func (h *WebAuthnHandler) FinishAuthentication(c *fiber.Ctx) error {
	// Get session
	session, err := h.sessionService.GetSession(c.Context(), c.Cookies("session"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	userID, err := primitive.ObjectIDFromHex(session.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user ID",
		})
	}

	// Get session data
	if session.WebAuthnData == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "no session data found",
		})
	}

	var sessionData webauthn.SessionData
	if err := json.Unmarshal([]byte(session.WebAuthnData), &sessionData); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to deserialize session data",
		})
	}

	// Parse response
	response, err := protocol.ParseCredentialRequestResponseBody(bytes.NewReader(c.Body()))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "failed to parse response",
		})
	}

	// Complete authentication
	if err := h.webAuthnService.FinishAuthentication(c.Context(), userID, sessionData, response); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Clear WebAuthn session data
	session.WebAuthnData = ""
	if err := h.sessionService.Update(c.Context(), session); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update session",
		})
	}

	return c.JSON(fiber.Map{
		"message": "authenticated successfully",
	})
}
