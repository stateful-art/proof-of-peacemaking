package handlers

import (
	"log"
	"proofofpeacemaking/internal/core/ports"

	"github.com/gofiber/fiber/v2"
)

type StatisticsHandler struct {
	statsService ports.StatisticsService
}

func NewStatisticsHandler(statsService ports.StatisticsService) *StatisticsHandler {
	return &StatisticsHandler{
		statsService: statsService,
	}
}

// ServeStatisticsPage renders the statistics page
func (h *StatisticsHandler) ServeStatisticsPage(c *fiber.Ctx) error {
	log.Printf("[STATISTICS] Fetching latest statistics")
	stats, err := h.statsService.GetLatestStats(c.Context())
	if err != nil {
		log.Printf("[STATISTICS] Error getting statistics: %v", err)
		stats = nil
	} else {
		log.Printf("[STATISTICS] Got statistics: Users=%d, Expressions=%d, Acks=%d",
			stats.TotalUsers, stats.TotalExpressions, stats.TotalAcknowledgements)
		log.Printf("[STATISTICS] Media stats: %v", stats.MediaStats)
		log.Printf("[STATISTICS] Citizenship stats: %v", stats.CitizenshipStats)
	}

	return c.Render("statistics", fiber.Map{
		"Title":      "Statistics",
		"Statistics": stats,
	})
}

// GetStatistics returns the latest statistics
func (h *StatisticsHandler) GetStatistics(c *fiber.Ctx) error {
	log.Printf("[STATISTICS] API: Fetching latest statistics")
	stats, err := h.statsService.GetLatestStats(c.Context())
	if err != nil {
		log.Printf("[STATISTICS] API: Error getting statistics: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get statistics",
		})
	}
	log.Printf("[STATISTICS] API: Returning statistics data")
	return c.JSON(stats)
}

// GetCountryList returns the list of available countries
func (h *StatisticsHandler) GetCountryList(c *fiber.Ctx) error {
	log.Printf("[STATISTICS] Fetching country list")
	countries, err := h.statsService.GetCountryList(c.Context())
	if err != nil {
		log.Printf("[STATISTICS] Error getting country list: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get country list",
		})
	}
	log.Printf("[STATISTICS] Returning %d countries", len(countries))
	return c.JSON(countries)
}

// UpdateStatistics triggers a statistics update
func (h *StatisticsHandler) UpdateStatistics(c *fiber.Ctx) error {
	log.Printf("[STATISTICS] Starting statistics update")
	if err := h.statsService.UpdateStats(c.Context()); err != nil {
		log.Printf("[STATISTICS] Error updating statistics: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update statistics",
		})
	}
	log.Printf("[STATISTICS] Statistics update completed successfully")
	return c.JSON(fiber.Map{
		"message": "Statistics updated successfully",
	})
}
