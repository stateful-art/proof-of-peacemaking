package services

import (
	"context"
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
	expressions, err := s.expressionService.List(ctx)
	if err != nil {
		return nil, err
	}

	var activities []map[string]interface{}
	for _, expr := range expressions {
		creator, err := s.userService.GetUserByAddress(ctx, expr.Creator.Hex())
		if err != nil {
			continue // Skip this expression if we can't get the creator
		}

		// Get acknowledgment count
		acks, err := s.acknowledgementService.ListByExpression(ctx, expr.ID.Hex())
		ackCount := 0
		if err == nil {
			ackCount = len(acks)
		}

		activities = append(activities, map[string]interface{}{
			"ID":                   expr.ID.Hex(),
			"UserAddress":          creator.Address,
			"Content":              expr.Content,
			"Timestamp":            expr.CreatedAt,
			"AcknowledgementCount": ackCount,
		})
	}

	return activities, nil
}
