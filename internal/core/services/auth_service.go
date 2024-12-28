package services

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type authService struct {
	userRepo  ports.UserRepository
	jwtSecret string
}

func NewAuthService(userRepo ports.UserRepository) ports.AuthService {
	// TODO: Load from config
	jwtSecret := "your-secret-key"
	return &authService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

func (s *authService) GenerateNonce(ctx context.Context, address string) (int, error) {
	log.Printf("[AUTH-SERVICE] Generating nonce for address: %s", address)
	nonce := rand.Intn(1000000)
	user, err := s.userRepo.FindByAddress(ctx, address)
	if err != nil {
		log.Printf("[AUTH-SERVICE] User not found, creating new user for address: %s", address)
		// If user doesn't exist, create one
		user = &domain.User{
			ID:        primitive.NewObjectID(),
			Address:   address,
			Nonce:     nonce, // Set initial nonce
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := s.userRepo.Create(ctx, user); err != nil {
			log.Printf("[AUTH-SERVICE] Error creating user: %v", err)
			return 0, fmt.Errorf("failed to create user: %w", err)
		}
		log.Printf("[AUTH-SERVICE] Successfully created new user with nonce: %d", nonce)
		return nonce, nil // Return nonce directly for new users
	}

	// Update nonce for existing user
	log.Printf("[AUTH-SERVICE] Updating nonce for existing user: %s", address)
	if err := s.userRepo.UpdateNonce(ctx, user.ID, nonce); err != nil {
		log.Printf("[AUTH-SERVICE] Error updating nonce: %v", err)
		return 0, fmt.Errorf("failed to update nonce: %w", err)
	}
	log.Printf("[AUTH-SERVICE] Successfully updated nonce to: %d", nonce)
	return nonce, nil
}

func (s *authService) generateToken(address string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"address": address,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // 24 hours
	})

	return token.SignedString([]byte(s.jwtSecret))
}

func (s *authService) VerifySignature(ctx context.Context, address string, signature string) (bool, string, error) {
	log.Printf("[AUTH-SERVICE] Verifying signature for address: %s", address)

	// Get user and their nonce
	user, err := s.userRepo.FindByAddress(ctx, address)
	if err != nil {
		log.Printf("[AUTH-SERVICE] Error finding user: %v", err)
		return false, "", fmt.Errorf("user not found: %w", err)
	}

	// Create the message that was signed
	message := fmt.Sprintf("Sign this message to verify your wallet. Nonce: %d", user.Nonce)
	log.Printf("[AUTH-SERVICE] Verifying message: %s", message)

	// Hash the message as Ethereum does
	hashedMessage := []byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message))
	messageHash := crypto.Keccak256Hash(hashedMessage)
	log.Printf("[AUTH-SERVICE] Message hash: %s", messageHash.Hex())

	// Decode signature
	decodedSig := hexutil.MustDecode(signature)
	log.Printf("[AUTH-SERVICE] Decoded signature length: %d", len(decodedSig))

	// The last byte is the recovery ID (v)
	v := decodedSig[64]
	log.Printf("[AUTH-SERVICE] Original v value: %d", v)

	// Convert recovery ID from MetaMask format
	if v >= 27 {
		v -= 27
	}
	decodedSig[64] = v
	log.Printf("[AUTH-SERVICE] Adjusted v value: %d", v)

	// Get public key
	sigPublicKeyECDSA, err := crypto.SigToPub(messageHash.Bytes(), decodedSig)
	if err != nil {
		log.Printf("[AUTH-SERVICE] Error recovering public key: %v", err)
		return false, "", fmt.Errorf("failed to recover public key: %w", err)
	}

	// Get address from public key
	recoveredAddr := crypto.PubkeyToAddress(*sigPublicKeyECDSA)
	log.Printf("[AUTH-SERVICE] Recovered address: %s", recoveredAddr.Hex())
	log.Printf("[AUTH-SERVICE] Expected address: %s", address)

	// Compare addresses (case-insensitive)
	if !strings.EqualFold(recoveredAddr.Hex(), address) {
		log.Printf("[AUTH-SERVICE] Address mismatch")
		return false, "", fmt.Errorf("signature does not match address")
	}

	log.Printf("[AUTH-SERVICE] Signature verified successfully")

	// Generate new nonce for next time
	newNonce := rand.Intn(1000000)
	if err := s.userRepo.UpdateNonce(ctx, user.ID, newNonce); err != nil {
		log.Printf("[AUTH-SERVICE] Error updating nonce: %v", err)
		return false, "", fmt.Errorf("failed to update nonce: %w", err)
	}
	log.Printf("[AUTH-SERVICE] Updated nonce to: %d", newNonce)

	// Generate JWT token
	token, err := s.generateToken(address)
	if err != nil {
		log.Printf("[AUTH-SERVICE] Error generating token: %v", err)
		return false, "", fmt.Errorf("failed to generate token: %w", err)
	}
	log.Printf("[AUTH-SERVICE] Successfully generated token")

	return true, token, nil
}

func (s *authService) Register(ctx context.Context, address string, email string) (*domain.User, string, error) {
	// Check if user exists
	user, err := s.userRepo.FindByAddress(ctx, address)
	if err == nil {
		// Update email if user exists
		user.Email = email
		user.UpdatedAt = time.Now()
		if err := s.userRepo.Update(ctx, user); err != nil {
			return nil, "", err
		}
	} else {
		// Create new user if doesn't exist
		user = &domain.User{
			ID:        primitive.NewObjectID(),
			Address:   address,
			Email:     email,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, "", err
		}
	}

	// Generate JWT token
	token, err := s.generateToken(address)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *authService) VerifyToken(ctx context.Context, tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if address, ok := claims["address"].(string); ok {
			return address, nil
		}
	}

	return "", fmt.Errorf("invalid token claims")
}
