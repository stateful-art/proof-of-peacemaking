package services

import (
	"context"
	"fmt"
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"
)

type acknowledgementService struct {
	acknowledgementRepo ports.AcknowledgementRepository
}

func NewAcknowledgementService(acknowledgementRepo ports.AcknowledgementRepository) ports.AcknowledgementService {
	return &acknowledgementService{
		acknowledgementRepo: acknowledgementRepo,
	}
}

func (s *acknowledgementService) Create(ctx context.Context, acknowledgement *domain.Acknowledgement) error {
	if err := s.acknowledgementRepo.Create(ctx, acknowledgement); err != nil {
		return fmt.Errorf("failed to create acknowledgement: %w", err)
	}
	return nil
}

func (s *acknowledgementService) ListByExpression(ctx context.Context, expressionID string) ([]*domain.Acknowledgement, error) {
	acknowledgements, err := s.acknowledgementRepo.FindByExpression(ctx, expressionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list acknowledgements by expression: %w", err)
	}
	return acknowledgements, nil
}

func (s *acknowledgementService) ListByUser(ctx context.Context, userAddress string) ([]*domain.Acknowledgement, error) {
	acknowledgements, err := s.acknowledgementRepo.FindByCreator(ctx, userAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to list acknowledgements by user: %w", err)
	}
	return acknowledgements, nil
}

func (s *acknowledgementService) Update(ctx context.Context, acknowledgement *domain.Acknowledgement) error {
	if err := s.acknowledgementRepo.Update(ctx, acknowledgement); err != nil {
		return fmt.Errorf("failed to update acknowledgement: %w", err)
	}
	return nil
}

func (s *acknowledgementService) ListByStatus(ctx context.Context, status domain.AcknowledgementStatus) ([]*domain.Acknowledgement, error) {
	acknowledgements, err := s.acknowledgementRepo.FindByStatus(ctx, status)
	if err != nil {
		return nil, fmt.Errorf("failed to list acknowledgements by status: %w", err)
	}
	return acknowledgements, nil
}
