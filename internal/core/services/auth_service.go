package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"
	"strings"
	"time"

	"log"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	userService ports.UserService
	sessionRepo ports.SessionRepository
}

func NewAuthService(userService ports.UserService, sessionRepo ports.SessionRepository) ports.AuthService {
	return &authService{
		userService: userService,
		sessionRepo: sessionRepo,
	}
}

func generateSecureToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (s *authService) GenerateNonce(ctx context.Context, address string) (int, error) {
	log.Printf("[AUTH] Generating nonce for address: %s", address)

	// Generate random nonce
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0, fmt.Errorf("failed to generate random nonce: %w", err)
	}
	nonce := int(n.Int64())

	// Find user
	user, err := s.userService.GetUserByAddress(ctx, address)
	if err != nil {
		log.Printf("[AUTH] Error finding user: %v", err)
		return 0, fmt.Errorf("failed to find user: %w", err)
	}

	// If user exists, update their nonce
	if user != nil {
		log.Printf("[AUTH] Found existing user with ID: %s", user.ID.Hex())
		if err := s.userService.UpdateNonce(ctx, user.ID, nonce); err != nil {
			log.Printf("[AUTH] Error updating nonce: %v", err)
			return 0, fmt.Errorf("failed to update nonce: %w", err)
		}
		log.Printf("[AUTH] Updated nonce to %d for user %s", nonce, user.ID.Hex())
	} else {
		// Create new user if not exists
		// Generate a valid username from the address (e.g., "0x844a54d19d")
		username := address[:10] // Take first 10 characters which includes "0x" and 8 hex chars

		user = &domain.User{
			ID:        primitive.NewObjectID(),
			Address:   address,
			Username:  username,
			Nonce:     nonce,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := s.userService.Create(ctx, user); err != nil {
			log.Printf("[AUTH] Error creating user: %v", err)
			return 0, fmt.Errorf("failed to create user: %w", err)
		}
		log.Printf("[AUTH] Created new user with ID: %s and nonce: %d", user.ID.Hex(), nonce)
	}

	return nonce, nil
}

func (s *authService) VerifySignature(ctx context.Context, address string, signature string) (bool, string, error) {
	log.Printf("[AUTH] Verifying signature for address: %s", address)

	// Get user and their nonce
	user, err := s.userService.GetUserByAddress(ctx, address)
	if err != nil {
		log.Printf("[AUTH] Error finding user: %v", err)
		return false, "", fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		log.Printf("[AUTH] User not found for address: %s", address)
		return false, "", fmt.Errorf("user not found")
	}
	log.Printf("[AUTH] Found user with ID: %s", user.ID.Hex())

	// Store current nonce for verification
	currentNonce := user.Nonce

	// Create the message that was signed
	message := fmt.Sprintf("Sign this message to verify your wallet. Nonce: %d", currentNonce)

	// Hash the message as Ethereum does
	fullMessage := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
	messageHash := crypto.Keccak256([]byte(fullMessage))

	// Decode signature
	decodedSig := hexutil.MustDecode(signature)
	if len(decodedSig) != 65 {
		return false, "", fmt.Errorf("invalid signature length")
	}

	// Extract r, s, v from signature
	signatureBytes := make([]byte, 65)
	copy(signatureBytes[0:32], decodedSig[0:32])
	copy(signatureBytes[32:64], decodedSig[32:64])

	// Adjust v if needed
	v := decodedSig[64]
	if v >= 27 {
		v -= 27
	}
	signatureBytes[64] = v

	// Recover public key
	pubKeyECDSA, err := crypto.SigToPub(messageHash, signatureBytes)
	if err != nil {
		log.Printf("[AUTH] Failed to recover public key: %v", err)
		return false, "", fmt.Errorf("failed to recover public key: %w", err)
	}

	// Get address from public key
	recoveredAddr := crypto.PubkeyToAddress(*pubKeyECDSA)

	// Compare addresses (case-insensitive)
	if !strings.EqualFold(recoveredAddr.Hex(), address) {
		log.Printf("[AUTH] Address mismatch - Recovered: %s, Expected: %s", recoveredAddr.Hex(), address)
		return false, "", fmt.Errorf("signature does not match address")
	}

	// Only update nonce after successful verification
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return false, "", fmt.Errorf("failed to generate new nonce: %w", err)
	}
	newNonce := int(n.Int64())
	if err := s.userService.UpdateNonce(ctx, user.ID, newNonce); err != nil {
		return false, "", fmt.Errorf("failed to update nonce: %w", err)
	}

	// Create session
	sessionToken, err := generateSecureToken()
	if err != nil {
		return false, "", fmt.Errorf("failed to generate session token: %w", err)
	}

	// Create session
	session := &domain.Session{
		ID:        primitive.NewObjectID(),
		UserID:    user.ID.Hex(),
		Token:     sessionToken,
		Address:   address,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return false, "", fmt.Errorf("failed to create session: %w", err)
	}

	log.Printf("[AUTH] Successfully verified signature and created session for address: %s", address)
	return true, sessionToken, nil
}

func (s *authService) Register(ctx context.Context, address string, email string) (*domain.User, string, error) {
	user, err := s.userService.GetUserByAddress(ctx, address)
	if err != nil {
		return nil, "", fmt.Errorf("failed to find user: %w", err)
	}

	if user == nil {
		user = &domain.User{
			ID:        primitive.NewObjectID(),
			Address:   address,
			Email:     email,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := s.userService.Create(ctx, user); err != nil {
			return nil, "", fmt.Errorf("failed to create user: %w", err)
		}
	} else {
		user.Email = email
		user.UpdatedAt = time.Now()
		if err := s.userService.Update(ctx, user); err != nil {
			return nil, "", fmt.Errorf("failed to update user: %w", err)
		}
	}

	// Create session
	sessionToken, err := generateSecureToken()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate session token: %w", err)
	}

	session := &domain.Session{
		ID:        primitive.NewObjectID(),
		UserID:    user.ID.Hex(),
		Token:     sessionToken,
		Address:   address,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, "", fmt.Errorf("failed to create session: %w", err)
	}

	return user, sessionToken, nil
}

func (s *authService) VerifyToken(ctx context.Context, token string) (string, error) {
	// log.Printf("[AUTH] Verifying token")
	session, err := s.sessionRepo.FindByToken(ctx, token)
	if err != nil {
		log.Printf("[AUTH] Error finding session: %v", err)
		return "", fmt.Errorf("failed to find session: %w", err)
	}
	if session == nil {
		log.Printf("[AUTH] Session not found")
		return "", fmt.Errorf("invalid or expired session")
	}
	if session.ExpiresAt.Before(time.Now()) {
		log.Printf("[AUTH] Session expired")
		return "", fmt.Errorf("session expired")
	}

	// Get user from session
	user, err := s.userService.GetUserByID(ctx, session.UserID)
	if err != nil {
		log.Printf("[AUTH] Error finding user: %v", err)
		return "", fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		log.Printf("[AUTH] User not found")
		return "", fmt.Errorf("invalid session: user not found")
	}

	// For wallet auth, return address
	if session.Address != "" {
		// log.Printf("[AUTH] Session verified for wallet address: %s", session.Address)
		return session.Address, nil
	}

	// For email auth, return email
	if user.Email != "" {
		// log.Printf("[AUTH] Session verified for email: %s", user.Email)
		return user.Email, nil
	}

	return "", fmt.Errorf("invalid session: no authentication method found")
}

func (s *authService) Logout(ctx context.Context, token string) error {
	// Find the session
	session, err := s.sessionRepo.FindByToken(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to find session: %w", err)
	}
	if session == nil {
		return nil // Session already invalid, nothing to do
	}

	// Invalidate the session by setting expiry to now
	session.ExpiresAt = time.Now()
	session.UpdatedAt = time.Now()

	// Update the session in the database
	if err := s.sessionRepo.Update(ctx, session); err != nil {
		return fmt.Errorf("failed to invalidate session: %w", err)
	}

	return nil
}

func (s *authService) RegisterWithEmail(ctx context.Context, email string, password string, username string) (*domain.User, string, error) {
	// Validate email format
	if !strings.Contains(email, "@") {
		return nil, "", fmt.Errorf("invalid email format")
	}

	// Validate username length
	if len(username) < 3 || len(username) > 30 {
		return nil, "", fmt.Errorf("username must be between 3 and 30 characters")
	}

	// Validate password strength
	if len(password) < 8 {
		return nil, "", fmt.Errorf("password must be at least 8 characters")
	}

	// Check if email already exists (case-insensitive)
	existingUser, err := s.userService.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, "", fmt.Errorf("error checking email: %w", err)
	}
	if existingUser != nil {
		return nil, "", fmt.Errorf("email already registered")
	}

	// Check if username already exists (case-insensitive)
	existingUser, err = s.userService.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, "", fmt.Errorf("error checking username: %w", err)
	}
	if existingUser != nil {
		return nil, "", fmt.Errorf("username already taken")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("error hashing password: %w", err)
	}

	// Create user
	user := &domain.User{
		ID:        primitive.NewObjectID(),
		Email:     strings.ToLower(email), // Store email in lowercase
		Username:  username,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.userService.Create(ctx, user); err != nil {
		return nil, "", fmt.Errorf("error creating user: %w", err)
	}

	// Create session
	sessionToken, err := generateSecureToken()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate session token: %w", err)
	}

	session := &domain.Session{
		ID:        primitive.NewObjectID(),
		UserID:    user.ID.Hex(),
		Token:     sessionToken,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, "", fmt.Errorf("failed to create session: %w", err)
	}

	return user, sessionToken, nil
}

func (s *authService) LoginWithEmail(ctx context.Context, email string, password string) (*domain.User, string, error) {
	// Get user by email
	user, err := s.userService.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, "", fmt.Errorf("invalid email or password")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, "", fmt.Errorf("invalid email or password")
	}

	// Create session
	sessionToken, err := generateSecureToken()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate session token")
	}

	session := &domain.Session{
		ID:        primitive.NewObjectID(),
		UserID:    user.ID.Hex(),
		Token:     sessionToken,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, "", fmt.Errorf("failed to create session")
	}

	return user, sessionToken, nil
}

func (s *authService) DeleteAllUserSessions(ctx context.Context, userIdentifier string) error {
	// Get user by email or address
	var user *domain.User
	var err error

	if strings.Contains(userIdentifier, "@") {
		user, err = s.userService.GetUserByEmail(ctx, userIdentifier)
	} else {
		user, err = s.userService.GetUserByAddress(ctx, userIdentifier)
	}

	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil // No sessions to delete if user doesn't exist
	}

	// Delete all sessions for this user
	if err := s.sessionRepo.DeleteByUserID(ctx, user.ID); err != nil {
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}

	return nil
}
