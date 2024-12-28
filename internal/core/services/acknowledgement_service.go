package services

import (
	"context"
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"
)

type acknowledgementService struct {
	userRepo ports.UserRepository
}

func NewAcknowledgementService(userRepo ports.UserRepository) ports.AcknowledgementService {
	return &acknowledgementService{userRepo: userRepo}
}

func (s *acknowledgementService) Create(ctx context.Context, userAddress string, expressionID string, content map[string]string) (*domain.Acknowledgement, error) {
	// TODO: Implement
	return nil, nil
}

func (s *acknowledgementService) ListByExpression(ctx context.Context, expressionID string) ([]*domain.Acknowledgement, error) {
	// TODO: Implement
	return nil, nil
}
