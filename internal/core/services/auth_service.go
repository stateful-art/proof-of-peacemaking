package services

import (
	"context"
	"math/rand"
	"proofofpeacemaking/internal/core/ports"
)

type authService struct {
	userRepo ports.UserRepository
}

func NewAuthService(userRepo ports.UserRepository) ports.AuthService {
	return &authService{userRepo: userRepo}
}

func (s *authService) GenerateNonce(ctx context.Context, address string) (int, error) {
	nonce := rand.Intn(1000000)
	user, err := s.userRepo.FindByAddress(ctx, address)
	if err != nil {
		return 0, err
	}
	return nonce, s.userRepo.UpdateNonce(ctx, user.ID, nonce)
}

func (s *authService) VerifySignature(ctx context.Context, address string, signature string) (bool, error) {
	// TODO: Implement actual signature verification
	return true, nil
}
