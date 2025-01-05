package mongodb

import (
	"context"
	"log"
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type StatisticsRepository struct {
	collection *mongo.Collection
}

func NewStatisticsRepository(db *mongo.Database) ports.StatisticsRepository {
	return &StatisticsRepository{
		collection: db.Collection("statistics"),
	}
}

func (r *StatisticsRepository) GetLatest(ctx context.Context) (*domain.Statistics, error) {
	log.Printf("[STATISTICS_REPO] Getting latest statistics")

	// Explicitly sort by createdAt in descending order (-1)
	opts := options.FindOne().SetSort(bson.D{{Key: "createdAt", Value: -1}})

	var stats domain.Statistics
	filter := bson.M{
		"createdAt": bson.M{
			"$exists": true, // ensure createdAt field exists
		},
	}

	err := r.collection.FindOne(ctx, filter, opts).Decode(&stats)
	if err == mongo.ErrNoDocuments {
		log.Printf("[STATISTICS_REPO] No statistics records found")
		// Return empty statistics if no records exist
		return &domain.Statistics{
			CitizenshipStats: make(map[string]int),
			MediaStats:       make(map[string]int),
		}, nil
	}
	if err != nil {
		log.Printf("[STATISTICS_REPO] Error retrieving statistics: %v", err)
		return nil, err
	}

	// Ensure maps are initialized
	if stats.CitizenshipStats == nil {
		stats.CitizenshipStats = make(map[string]int)
	}
	if stats.MediaStats == nil {
		stats.MediaStats = make(map[string]int)
	}

	log.Printf("[STATISTICS_REPO] Retrieved statistics: Users=%d, Expressions=%d, Acks=%d",
		stats.TotalUsers, stats.TotalExpressions, stats.TotalAcknowledgements)
	log.Printf("[STATISTICS_REPO] CitizenshipStats: %v", stats.CitizenshipStats)
	log.Printf("[STATISTICS_REPO] MediaStats: %v", stats.MediaStats)
	return &stats, nil
}

func (r *StatisticsRepository) Create(ctx context.Context, stats *domain.Statistics) error {
	log.Printf("[STATISTICS_REPO] Creating new statistics record")

	// Ensure createdAt is set to current time
	stats.CreatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, stats)
	if err != nil {
		log.Printf("[STATISTICS_REPO] Error creating statistics record: %v", err)
		return err
	}
	log.Printf("[STATISTICS_REPO] Statistics record created successfully")
	return nil
}

func (r *StatisticsRepository) GetCountryList(ctx context.Context) ([]domain.CountryInfo, error) {
	log.Printf("[STATISTICS_REPO] Getting country list")
	countries := domain.GetCountryList()
	log.Printf("[STATISTICS_REPO] Retrieved %d countries", len(countries))
	return countries, nil
}
