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

	// Create a new user with pending status
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
		// If registration fails, we should clean up the user
		if delErr := h.userService.Delete(c.Context(), user.ID); delErr != nil {
			log.Printf("Failed to delete user after failed registration: %v", delErr)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Create a temporary registration session
	session := &domain.Session{
		UserID:         user.ID.Hex(),
		WebAuthnData:   string(must(json.Marshal(sessionData))),
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(5 * time.Minute), // Short expiry for registration
		IsRegistration: true,                            // Mark this as a registration session
	}

	if err := h.sessionService.Create(c.Context(), session); err != nil {
		// If session creation fails, clean up the user
		if delErr := h.userService.Delete(c.Context(), user.ID); delErr != nil {
			log.Printf("Failed to delete user after failed session creation: %v", delErr)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create session",
		})
	}

	// Set temporary registration session cookie
	c.Cookie(&fiber.Cookie{
		Name:     "registration_session",
		Value:    session.Token,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
		MaxAge:   300, // 5 minutes
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
	// Get user from registration session
	session, err := h.sessionService.GetSession(c.Context(), c.Cookies("registration_session"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	if !session.IsRegistration {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid session type",
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

	// Delete the registration session
	if err := h.sessionService.Delete(c.Context(), session.Token); err != nil {
		log.Printf("Failed to delete registration session: %v", err)
	}

	// Create a new authenticated session
	authSession := &domain.Session{
		UserID:    userID.Hex(),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24-hour session
	}

	if err := h.sessionService.Create(c.Context(), authSession); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create authenticated session",
		})
	}

	// Set authenticated session cookie
	c.Cookie(&fiber.Cookie{
		Name:     "session",
		Value:    authSession.Token,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
		MaxAge:   86400, // 24 hours
	})

	return c.JSON(fiber.Map{
		"message": "passkey registered successfully",
	})
}

// BeginAuthentication initiates the passkey authentication process
func (h *WebAuthnHandler) BeginAuthentication(c *fiber.Ctx) error {
	// Parse request body to get email
	var req struct {
		Email string `json:"email"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "email is required",
		})
	}

	// Get user by email
	user, err := h.userService.GetUserByEmail(c.Context(), req.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get user",
		})
	}
	if user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "user not found",
		})
	}

	// Begin authentication
	options, sessionData, err := h.webAuthnService.BeginAuthentication(c.Context(), user.ID)
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
		UserID:       user.ID.Hex(),
		WebAuthnData: string(sessionDataJSON),
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(5 * time.Minute), // Short expiry for authentication
	}

	if err := h.sessionService.Create(c.Context(), session); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to store session data",
		})
	}

	// Set session cookie
	c.Cookie(&fiber.Cookie{
		Name:     "auth_session",
		Value:    session.Token,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
		MaxAge:   300, // 5 minutes
	})

	return c.JSON(options)
}

// FinishAuthentication completes the passkey authentication process
func (h *WebAuthnHandler) FinishAuthentication(c *fiber.Ctx) error {
	log.Printf("[WEBAUTHN] Starting FinishAuthentication")

	// Get session
	session, err := h.sessionService.GetSession(c.Context(), c.Cookies("auth_session"))
	if err != nil {
		log.Printf("[WEBAUTHN] Failed to get auth session: %v", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}
	log.Printf("[WEBAUTHN] Found auth session for user: %s", session.UserID)

	userID, err := primitive.ObjectIDFromHex(session.UserID)
	if err != nil {
		log.Printf("[WEBAUTHN] Invalid user ID format: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user ID",
		})
	}

	// Get session data
	if session.WebAuthnData == "" {
		log.Printf("[WEBAUTHN] No WebAuthn data found in session")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "no session data found",
		})
	}

	var sessionData webauthn.SessionData
	if err := json.Unmarshal([]byte(session.WebAuthnData), &sessionData); err != nil {
		log.Printf("[WEBAUTHN] Failed to deserialize session data: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to deserialize session data",
		})
	}
	log.Printf("[WEBAUTHN] Successfully deserialized session data")

	// Log request body for debugging
	body := c.Body()

	// Parse response
	response, err := protocol.ParseCredentialRequestResponseBody(bytes.NewReader(body))
	if err != nil {
		log.Printf("[WEBAUTHN] Failed to parse credential response: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "failed to parse response",
		})
	}
	log.Printf("[WEBAUTHN] Successfully parsed credential response")

	// Complete authentication
	if err := h.webAuthnService.FinishAuthentication(c.Context(), userID, sessionData, response); err != nil {
		log.Printf("[WEBAUTHN] Failed to finish authentication: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	log.Printf("[WEBAUTHN] Successfully finished authentication")

	// Delete the temporary auth session
	if err := h.sessionService.Delete(c.Context(), session.Token); err != nil {
		log.Printf("[WEBAUTHN] Failed to delete auth session: %v", err)
	}

	// Create a new authenticated session
	authSession := &domain.Session{
		UserID:    userID.Hex(),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24-hour session
	}

	if err := h.sessionService.Create(c.Context(), authSession); err != nil {
		log.Printf("[WEBAUTHN] Failed to create authenticated session: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create authenticated session",
		})
	}
	log.Printf("[WEBAUTHN] Created new authenticated session")

	// Set authenticated session cookie
	c.Cookie(&fiber.Cookie{
		Name:     "session",
		Value:    authSession.Token,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
		MaxAge:   86400, // 24 hours
	})
	log.Printf("[WEBAUTHN] Set session cookie")

	return c.JSON(fiber.Map{
		"message": "authenticated successfully",
	})
}
