package services

import (
	"context"
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"
)

type expressionService struct {
	userRepo ports.UserRepository
}

func NewExpressionService(userRepo ports.UserRepository) ports.ExpressionService {
	return &expressionService{userRepo: userRepo}
}

func (s *expressionService) Create(ctx context.Context, userAddress string, content map[string]string) (*domain.Expression, error) {
	// TODO: Implement
	return nil, nil
}

func (s *expressionService) List(ctx context.Context) ([]*domain.Expression, error) {
	// TODO: Implement
	return nil, nil
}

func (s *expressionService) Get(ctx context.Context, id string) (*domain.Expression, error) {
	// TODO: Implement
	return nil, nil
}
