package services

import (
	"context"
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"
)

type proofNFTService struct {
	userRepo ports.UserRepository
}

func NewProofNFTService(userRepo ports.UserRepository) ports.ProofNFTService {
	return &proofNFTService{userRepo: userRepo}
}

func (s *proofNFTService) RequestProof(ctx context.Context, expressionID string, acknowledgementID string) error {
	// TODO: Implement
	return nil
}

func (s *proofNFTService) ApproveProof(ctx context.Context, requestID string) error {
	// TODO: Implement
	return nil
}

func (s *proofNFTService) ListUserProofs(ctx context.Context, userAddress string) ([]*domain.ProofNFT, error) {
	// TODO: Implement
	return nil, nil
}
