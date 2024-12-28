package handlers

import (
	"proofofpeacemaking/internal/core/ports"

	"log"

	"github.com/gofiber/fiber/v2"
)

type DashboardHandler struct {
	expressionService      ports.ExpressionService
	acknowledgementService ports.AcknowledgementService
	userService            ports.UserService
	proofNFTService        ports.ProofNFTService
}

func NewDashboardHandler(
	expressionService ports.ExpressionService,
	acknowledgementService ports.AcknowledgementService,
	userService ports.UserService,
	proofNFTService ports.ProofNFTService,
) *DashboardHandler {
	return &DashboardHandler{
		expressionService:      expressionService,
		acknowledgementService: acknowledgementService,
		userService:            userService,
		proofNFTService:        proofNFTService,
	}
}

func (h *DashboardHandler) GetDashboard(c *fiber.Ctx) error {
	log.Printf("[DASHBOARD] Starting dashboard handler")

	// Get user address from context (set by auth middleware)
	userAddress, ok := c.Locals("userAddress").(string)
	if !ok {
		log.Printf("[DASHBOARD] Error: User address not found in context")
		// Redirect to home page if not authenticated
		return c.Redirect("/")
	}
	log.Printf("[DASHBOARD] Got user address: %s", userAddress)

	// Get user's expressions
	expressions, err := h.expressionService.ListByUser(c.Context(), userAddress)
	if err != nil {
		log.Printf("[DASHBOARD] Error fetching expressions: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch expressions",
		})
	}

	// Get user's proofs
	proofs, err := h.proofNFTService.ListUserProofs(c.Context(), userAddress)
	if err != nil {
		log.Printf("[DASHBOARD] Error fetching proofs: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch proofs",
		})
	}

	// Get user's acknowledgements
	acknowledgements, err := h.acknowledgementService.ListByUser(c.Context(), userAddress)
	if err != nil {
		log.Printf("[DASHBOARD] Error fetching acknowledgements: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch acknowledgements",
		})
	}

	// Get user's stats
	stats := fiber.Map{
		"TotalExpressions":      len(expressions),
		"TotalAcknowledgements": len(acknowledgements),
		"TotalProofs":           len(proofs),
	}

	// Sort expressions by timestamp to get most recent
	// TODO: Add proper sorting when timestamp field is added
	recentExpressions := expressions
	if len(recentExpressions) > 5 {
		recentExpressions = recentExpressions[:5]
	}

	// Sort proofs by timestamp to get most recent
	recentProofs := proofs
	if len(recentProofs) > 5 {
		recentProofs = recentProofs[:5]
	}

	data := fiber.Map{
		"Title":             "Dashboard - Proof of Peacemaking",
		"User":              fiber.Map{"Address": userAddress},
		"Stats":             stats,
		"RecentExpressions": recentExpressions,
		"Proofs":            recentProofs,
	}
	log.Printf("[DASHBOARD] Data being passed to template: %+v", data)

	return c.Render("dashboard", data, "")
}
