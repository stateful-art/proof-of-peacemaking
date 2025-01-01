package services

import (
	"context"
	"fmt"
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type proofNFTService struct {
	userRepo     ports.UserRepository
	proofNFTRepo ports.ProofNFTRepository
}

func NewProofNFTService(userRepo ports.UserRepository, proofNFTRepo ports.ProofNFTRepository) ports.ProofNFTService {
	return &proofNFTService{
		userRepo:     userRepo,
		proofNFTRepo: proofNFTRepo,
	}
}

func (s *proofNFTService) RequestProof(ctx context.Context, expressionID string, acknowledgementID string) error {
	// Get user from context
	user, err := s.userRepo.GetByAddress(ctx, expressionID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	// Create new ProofNFT
	proofNFT := &domain.ProofNFT{
		ID:           primitive.NewObjectID(),
		TokenID:      0, // Will be set by smart contract
		Expression:   expressionID,
		Acknowledger: user.ID.Hex(),
		Status:       string(domain.ProofRequestPending),
		CreatedAt:    time.Now(),
	}

	// Save to database
	if err := s.proofNFTRepo.Create(ctx, proofNFT); err != nil {
		return fmt.Errorf("failed to create proof NFT: %w", err)
	}

	return nil
}

func (s *proofNFTService) ApproveProof(ctx context.Context, requestID string) error {
	// Get user from context
	user, err := s.userRepo.GetByAddress(ctx, requestID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	// TODO: Implement proof approval logic
	return nil
}

func (s *proofNFTService) ListUserProofs(ctx context.Context, userAddress string) ([]*domain.ProofNFT, error) {
	// Get user by address
	user, err := s.userRepo.GetByAddress(ctx, userAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// TODO: Implement listing user proofs
	return nil, nil
}
