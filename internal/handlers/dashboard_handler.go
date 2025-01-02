package handlers

import (
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"
	"strings"

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

	// Get user identifier from context (set by auth middleware)
	userIdentifier, ok := c.Locals("userAddress").(string)
	if !ok {
		log.Printf("[DASHBOARD] Error: User identifier not found in context")
		// Redirect to home page if not authenticated
		return c.Redirect("/")
	}
	log.Printf("[DASHBOARD] Got user identifier: %s", userIdentifier)

	// Get user by email or address
	var user *domain.User
	var err error
	if strings.Contains(userIdentifier, "@") {
		user, err = h.userService.GetUserByEmail(c.Context(), userIdentifier)
	} else {
		user, err = h.userService.GetUserByAddress(c.Context(), userIdentifier)
	}

	if err != nil {
		log.Printf("[DASHBOARD] Error getting user: %v", err)
		// Instead of returning error, show empty dashboard
		user = &domain.User{
			Email: userIdentifier,
		}
	}

	// Initialize empty slices for data
	var expressions []*domain.Expression
	var proofs []*domain.ProofNFT
	var acknowledgements []*domain.Acknowledgement

	// Get user's expressions
	if user != nil {
		expressions, err = h.expressionService.ListByUser(c.Context(), userIdentifier)
		if err != nil {
			log.Printf("[DASHBOARD] Error fetching expressions: %v", err)
			// Continue with empty expressions
			expressions = []*domain.Expression{}
		}

		// Get user's proofs
		proofs, err = h.proofNFTService.ListUserProofs(c.Context(), userIdentifier)
		if err != nil {
			log.Printf("[DASHBOARD] Error fetching proofs: %v", err)
			// Continue with empty proofs
			proofs = []*domain.ProofNFT{}
		}

		// Get user's acknowledgements
		acknowledgements, err = h.acknowledgementService.ListByUser(c.Context(), userIdentifier)
		if err != nil {
			log.Printf("[DASHBOARD] Error fetching acknowledgements: %v", err)
			// Continue with empty acknowledgements
			acknowledgements = []*domain.Acknowledgement{}
		}
	}

	// Get user's stats
	stats := fiber.Map{
		"TotalExpressions":      len(expressions),
		"TotalAcknowledgements": len(acknowledgements),
		"TotalProofs":           len(proofs),
	}

	// Sort expressions by timestamp to get most recent
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
		"Title": "Dashboard - Proof of Peacemaking",
		"User": fiber.Map{
			"Email":   user.Email,
			"Address": user.Address,
		},
		"Stats":             stats,
		"RecentExpressions": recentExpressions,
		"Proofs":            recentProofs,
	}
	log.Printf("[DASHBOARD] Data being passed to template: %+v", data)

	return c.Render("dashboard", data, "")
}
