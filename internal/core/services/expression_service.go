package services

import (
	"context"
	"fmt"
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"
)

type expressionService struct {
	expressionRepo      ports.ExpressionRepository
	acknowledgementRepo ports.AcknowledgementRepository
}

func NewExpressionService(expressionRepo ports.ExpressionRepository, acknowledgementRepo ports.AcknowledgementRepository) ports.ExpressionService {
	return &expressionService{
		expressionRepo:      expressionRepo,
		acknowledgementRepo: acknowledgementRepo,
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
	if expression == nil {
		return nil, nil
	}

	// Initialize counts to 0
	expression.ActiveAcknowledgementCount = 0
	expression.Acknowledgements = []*domain.Acknowledgement{}

	// Get acknowledgements for the expression
	acks, err := s.acknowledgementRepo.FindByExpression(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get acknowledgements for expression: %w", err)
	}
	expression.Acknowledgements = acks

	// Calculate active acknowledgement count
	for _, ack := range acks {
		if ack.Status == domain.AcknowledgementStatusActive {
			expression.ActiveAcknowledgementCount++
		}
	}

	return expression, nil
}

func (s *expressionService) List(ctx context.Context) ([]*domain.Expression, error) {
	expressions, err := s.expressionRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list expressions: %w", err)
	}

	// For each expression, get its acknowledgements and calculate counts
	for _, expr := range expressions {
		// Initialize counts to 0
		expr.ActiveAcknowledgementCount = 0
		expr.Acknowledgements = []*domain.Acknowledgement{}

		// Get acknowledgements
		acks, err := s.acknowledgementRepo.FindByExpression(ctx, expr.ID.Hex())
		if err != nil {
			return nil, fmt.Errorf("failed to get acknowledgements for expression: %w", err)
		}
		expr.Acknowledgements = acks

		// Calculate active acknowledgement count
		for _, ack := range acks {
			if ack.Status == domain.AcknowledgementStatusActive {
				expr.ActiveAcknowledgementCount++
			}
		}
	}

	return expressions, nil
}

func (s *expressionService) ListByUser(ctx context.Context, userAddress string) ([]*domain.Expression, error) {
	expressions, err := s.expressionRepo.FindByCreator(ctx, userAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to list expressions by user: %w", err)
	}

	// For each expression, get its acknowledgements and calculate counts
	for _, expr := range expressions {
		// Initialize counts to 0
		expr.ActiveAcknowledgementCount = 0
		expr.Acknowledgements = []*domain.Acknowledgement{}

		// Get acknowledgements
		acks, err := s.acknowledgementRepo.FindByExpression(ctx, expr.ID.Hex())
		if err != nil {
			return nil, fmt.Errorf("failed to get acknowledgements for expression: %w", err)
		}
		expr.Acknowledgements = acks

		// Calculate active acknowledgement count
		for _, ack := range acks {
			if ack.Status == domain.AcknowledgementStatusActive {
				expr.ActiveAcknowledgementCount++
			}
		}
	}

	return expressions, nil
}
