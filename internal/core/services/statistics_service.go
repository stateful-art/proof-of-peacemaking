package services

import (
	"context"
	"log"
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"
)

type statisticsService struct {
	statsRepo      ports.StatisticsRepository
	userRepo       ports.UserRepository
	expressionRepo ports.ExpressionRepository
}

func NewStatisticsService(
	statsRepo ports.StatisticsRepository,
	userRepo ports.UserRepository,
	expressionRepo ports.ExpressionRepository,
) ports.StatisticsService {
	return &statisticsService{
		statsRepo:      statsRepo,
		userRepo:       userRepo,
		expressionRepo: expressionRepo,
	}
}

// GetLatestStats returns the most recent statistics
func (s *statisticsService) GetLatestStats(ctx context.Context) (*domain.Statistics, error) {
	log.Printf("[STATISTICS_SERVICE] Getting latest statistics")
	stats, err := s.statsRepo.GetLatest(ctx)
	if err != nil {
		log.Printf("[STATISTICS_SERVICE] Error getting latest statistics: %v", err)
		return nil, err
	}
	if stats != nil {
		log.Printf("[STATISTICS_SERVICE] Retrieved statistics: Users=%d, Expressions=%d, Acks=%d",
			stats.TotalUsers, stats.TotalExpressions, stats.TotalAcknowledgements)
	} else {
		log.Printf("[STATISTICS_SERVICE] No statistics found, will return nil")
	}
	return stats, nil
}

// UpdateStats creates a new statistics record
func (s *statisticsService) UpdateStats(ctx context.Context) error {
	log.Printf("[STATISTICS_SERVICE] Starting statistics update")

	// Get total users
	totalUsers, err := s.userRepo.GetTotalCount(ctx)
	if err != nil {
		log.Printf("[STATISTICS_SERVICE] Error getting total users: %v", err)
		return err
	}
	log.Printf("[STATISTICS_SERVICE] Total users: %d", totalUsers)

	// Get total expressions
	totalExpressions, err := s.expressionRepo.GetTotalCount(ctx)
	if err != nil {
		log.Printf("[STATISTICS_SERVICE] Error getting total expressions: %v", err)
		return err
	}
	log.Printf("[STATISTICS_SERVICE] Total expressions: %d", totalExpressions)

	// Get total acknowledgements
	totalAcknowledgements, err := s.expressionRepo.GetTotalAcknowledgements(ctx)
	if err != nil {
		log.Printf("[STATISTICS_SERVICE] Error getting total acknowledgements: %v", err)
		return err
	}
	log.Printf("[STATISTICS_SERVICE] Total acknowledgements: %d", totalAcknowledgements)

	// Get citizenship distribution
	citizenshipStats, err := s.userRepo.GetCitizenshipDistribution(ctx)
	if err != nil {
		log.Printf("[STATISTICS_SERVICE] Error getting citizenship distribution: %v", err)
		return err
	}
	if len(citizenshipStats) == 0 {
		citizenshipStats = map[string]int{
			"UNKNOWN": totalUsers, // Default all users to unknown if no citizenship data
		}
	}
	log.Printf("[STATISTICS_SERVICE] Citizenship stats: %v", citizenshipStats)

	// Get media type distribution
	mediaStats, err := s.expressionRepo.GetMediaTypeDistribution(ctx)
	if err != nil {
		log.Printf("[STATISTICS_SERVICE] Error getting media distribution: %v", err)
		return err
	}
	if len(mediaStats) == 0 {
		mediaStats = map[string]int{
			"text": totalExpressions, // Default all expressions to text if no media type data
		}
	}
	log.Printf("[STATISTICS_SERVICE] Media stats: %v", mediaStats)

	// Create new statistics record
	stats := &domain.Statistics{
		TotalUsers:            totalUsers,
		TotalExpressions:      totalExpressions,
		TotalAcknowledgements: totalAcknowledgements,
		CitizenshipStats:      citizenshipStats,
		MediaStats:            mediaStats,
	}

	log.Printf("[STATISTICS_SERVICE] Creating new statistics record")
	err = s.statsRepo.Create(ctx, stats)
	if err != nil {
		log.Printf("[STATISTICS_SERVICE] Error creating statistics record: %v", err)
		return err
	}
	log.Printf("[STATISTICS_SERVICE] Statistics update completed successfully")
	return nil
}

// GetCountryList returns available countries for citizenship
func (s *statisticsService) GetCountryList(ctx context.Context) ([]domain.CountryInfo, error) {
	log.Printf("[STATISTICS_SERVICE] Getting country list")
	countries, err := s.statsRepo.GetCountryList(ctx)
	if err != nil {
		log.Printf("[STATISTICS_SERVICE] Error getting country list: %v", err)
		return nil, err
	}
	log.Printf("[STATISTICS_SERVICE] Retrieved %d countries", len(countries))
	return countries, nil
}

func (s *statisticsService) UpdateStatisticsAfterExpression(ctx context.Context) error {
	log.Printf("[STATISTICS_SERVICE] Updating statistics after new expression")
	return s.UpdateStats(ctx)
}

func (s *statisticsService) UpdateStatisticsAfterAcknowledgement(ctx context.Context) error {
	log.Printf("[STATISTICS_SERVICE] Updating statistics after new acknowledgement")
	return s.UpdateStats(ctx)
}

func (s *statisticsService) UpdateStatisticsAfterCitizenshipChange(ctx context.Context) error {
	log.Printf("[STATISTICS_SERVICE] Updating statistics after citizenship change")
	return s.UpdateStats(ctx)
}
