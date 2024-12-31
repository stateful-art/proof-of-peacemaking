package services

import (
	"context"
	"log"
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
		log.Printf("[FEED] Error listing expressions: %v", err)
		return nil, err
	}

	var activities []map[string]interface{}
	for _, expr := range expressions {
		if expr == nil {
			log.Printf("[FEED] Skipping nil expression")
			continue
		}

		// Get acknowledgment count
		acks, err := s.acknowledgementService.ListByExpression(ctx, expr.ID.Hex())
		ackCount := 0
		if err != nil {
			log.Printf("[FEED] Error getting acknowledgments for expression %s: %v", expr.ID.Hex(), err)
		} else {
			ackCount = len(acks)
		}

		activity := map[string]interface{}{
			"ID":                   expr.ID.Hex(),
			"UserAddress":          expr.CreatorAddress,
			"Content":              expr.Content,
			"Timestamp":            expr.CreatedAt,
			"AcknowledgementCount": ackCount,
		}

		activities = append(activities, activity)
	}

	if len(activities) == 0 {
		log.Printf("[FEED] No activities found")
	} else {
		log.Printf("[FEED] Found %d activities", len(activities))
	}

	return activities, nil
}
