package handlers

import (
	"log"
	"proofofpeacemaking/internal/core/ports"

	"github.com/gofiber/fiber/v2"
)

type CountryHandler struct {
	countryService ports.CountryService
}

func NewCountryHandler(countryService ports.CountryService) *CountryHandler {
	return &CountryHandler{
		countryService: countryService,
	}
}

// SearchCountries handles country search requests
func (h *CountryHandler) SearchCountries(c *fiber.Ctx) error {
	query := c.Query("search", "")
	log.Printf("[COUNTRY] Searching countries with query: %s", query)

	countries, err := h.countryService.SearchCountries(c.Context(), query)
	if err != nil {
		log.Printf("[COUNTRY] Error searching countries: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to search countries",
		})
	}

	log.Printf("[COUNTRY] Found %d matches", len(countries))
	return c.JSON(countries)
}
