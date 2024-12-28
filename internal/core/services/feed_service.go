package services

import (
	"context"
	"proofofpeacemaking/internal/core/ports"
)

type feedService struct {
	expressionService ports.ExpressionService
	userService       ports.UserService
}

func NewFeedService(expressionService ports.ExpressionService, userService ports.UserService) ports.FeedService {
	return &feedService{
		expressionService: expressionService,
		userService:       userService,
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

		activities = append(activities, map[string]interface{}{
			"UserAddress": creator.Address,
			"Content":     expr.Content,
			"Timestamp":   expr.CreatedAt,
		})
	}

	return activities, nil
}
