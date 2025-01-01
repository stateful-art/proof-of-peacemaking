package ports

import (
	"context"
	"proofofpeacemaking/internal/core/domain"
)

type ProofNFTRepository interface {
	Create(ctx context.Context, proofNFT *domain.ProofNFT) error
	Update(ctx context.Context, proofNFT *domain.ProofNFT) error
	FindByID(ctx context.Context, id string) (*domain.ProofNFT, error)
	FindByAcknowledger(ctx context.Context, acknowledgerID string) ([]*domain.ProofNFT, error)
}
