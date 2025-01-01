package services

import (
	"context"
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"
)

type feedService struct {
	expressionService      ports.ExpressionService
	userService            ports.UserService
	acknowledgementService ports.AcknowledgementService
}

func NewFeedService(
	expressionService ports.ExpressionService,
	userService ports.UserService,
	acknowledgementService ports.AcknowledgementService,
) ports.FeedService {
	return &feedService{
		expressionService:      expressionService,
		userService:            userService,
		acknowledgementService: acknowledgementService,
	}
}

func (s *feedService) GetFeed(ctx context.Context) ([]map[string]interface{}, error) {
	// Get all expressions
	expressions, err := s.expressionService.List(ctx)
	if err != nil {
		return nil, err
	}

	// Convert expressions to feed items
	var feedItems []map[string]interface{}
	for _, expr := range expressions {
		// Get acknowledgements for this expression
		acks, err := s.acknowledgementService.ListByExpression(ctx, expr.ID.Hex())
		if err != nil {
			return nil, err
		}

		// Count active acknowledgements
		activeCount := 0
		for _, ack := range acks {
			if ack.Status == domain.AcknowledgementStatusActive {
				activeCount++
			}
		}

		// Convert ObjectID to hex string for template
		feedItem := map[string]interface{}{
			"ID":                         expr.ID.Hex(),
			"CreatorAddress":             expr.CreatorAddress,
			"Content":                    expr.Content,
			"Timestamp":                  expr.CreatedAt,
			"Acknowledgements":           acks,
			"ActiveAcknowledgementCount": activeCount,
		}
		feedItems = append(feedItems, feedItem)
	}

	return feedItems, nil
}
