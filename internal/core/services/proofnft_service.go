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
	user, err := s.userRepo.FindByAddress(ctx, expressionID)
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

	// Save to repository
	if err := s.proofNFTRepo.Create(ctx, proofNFT); err != nil {
		return fmt.Errorf("failed to create proof NFT: %w", err)
	}

	return nil
}

func (s *proofNFTService) ApproveProof(ctx context.Context, requestID string) error {
	// Get the proof request
	proofNFT, err := s.proofNFTRepo.FindByID(ctx, requestID)
	if err != nil {
		return fmt.Errorf("failed to find proof NFT: %w", err)
	}
	if proofNFT == nil {
		return fmt.Errorf("proof NFT not found")
	}

	// Update status
	proofNFT.Status = string(domain.ProofRequestAccepted)
	now := time.Now()
	proofNFT.MintedAt = &now

	// Save changes
	if err := s.proofNFTRepo.Update(ctx, proofNFT); err != nil {
		return fmt.Errorf("failed to update proof NFT: %w", err)
	}

	return nil
}

func (s *proofNFTService) ListUserProofs(ctx context.Context, userAddress string) ([]*domain.ProofNFT, error) {
	// Get user
	user, err := s.userRepo.FindByAddress(ctx, userAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Get proofs for user
	proofs, err := s.proofNFTRepo.FindByAcknowledger(ctx, user.ID.Hex())
	if err != nil {
		return nil, fmt.Errorf("failed to find proofs: %w", err)
	}

	return proofs, nil
}
