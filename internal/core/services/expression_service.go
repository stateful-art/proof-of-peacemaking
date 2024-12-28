package services

import (
	"context"
	"fmt"
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"
)

type expressionService struct {
	expressionRepo ports.ExpressionRepository
}

func NewExpressionService(expressionRepo ports.ExpressionRepository) ports.ExpressionService {
	return &expressionService{
		expressionRepo: expressionRepo,
	}
}

func (s *expressionService) Create(ctx context.Context, expression *domain.Expression) error {
	if err := s.expressionRepo.Create(ctx, expression); err != nil {
		return fmt.Errorf("failed to create expression: %w", err)
	}
	return nil
}

func (s *expressionService) Get(ctx context.Context, id string) (*domain.Expression, error) {
	expression, err := s.expressionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get expression: %w", err)
	}
	return expression, nil
}

func (s *expressionService) List(ctx context.Context) ([]*domain.Expression, error) {
	expressions, err := s.expressionRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list expressions: %w", err)
	}
	return expressions, nil
}

func (s *expressionService) ListByUser(ctx context.Context, userAddress string) ([]*domain.Expression, error) {
	expressions, err := s.expressionRepo.FindByCreator(ctx, userAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch expressions by user: %w", err)
	}
	return expressions, nil
}
