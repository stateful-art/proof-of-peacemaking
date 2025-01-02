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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user data",
		})
	}

	// Get user's expressions
	expressions, err := h.expressionService.ListByUser(c.Context(), user.ID.Hex())
	if err != nil {
		log.Printf("[DASHBOARD] Error fetching expressions: %v", err)
		expressions = []*domain.Expression{} // Use empty slice instead of failing
	}

	// Get acknowledgments received by user's expressions
	totalAcksReceived := 0
	uniqueAcknowledgers := make(map[string]bool)
	for _, expr := range expressions {
		acks, err := h.acknowledgementService.ListByExpression(c.Context(), expr.ID.Hex())
		if err != nil {
			log.Printf("[DASHBOARD] Error fetching acks for expression %s: %v", expr.ID.Hex(), err)
			continue
		}
		for _, ack := range acks {
			if ack.Status == domain.AcknowledgementStatusActive {
				totalAcksReceived++
				uniqueAcknowledgers[ack.Acknowledger] = true
			}
		}
	}

	// Get acknowledgments made by user
	acknowledgementsMade, err := h.acknowledgementService.ListByUser(c.Context(), user.ID.Hex())
	if err != nil {
		log.Printf("[DASHBOARD] Error fetching acknowledgements: %v", err)
		acknowledgementsMade = []*domain.Acknowledgement{} // Use empty slice instead of failing
	}

	// Count acknowledgments by status
	activeAcks := 0
	refutedAcks := 0
	uniqueExpressionsAcked := make(map[string]bool)
	uniqueCreatorsAcked := make(map[string]bool)

	for _, ack := range acknowledgementsMade {
		uniqueExpressionsAcked[ack.ExpressionID] = true

		// Get expression to find its creator
		expr, err := h.expressionService.Get(c.Context(), ack.ExpressionID)
		if err != nil {
			log.Printf("[DASHBOARD] Error fetching expression %s: %v", ack.ExpressionID, err)
			continue
		}
		uniqueCreatorsAcked[expr.Creator] = true

		// Count by status
		if ack.Status == domain.AcknowledgementStatusActive {
			activeAcks++
		} else if ack.Status == domain.AcknowledgementStatusRefuted {
			refutedAcks++
		}
	}

	// Prepare stats
	expressionStats := fiber.Map{
		"TotalExpressions":      len(expressions),
		"TotalAcknowledgements": totalAcksReceived,
		"UniqueAcknowledgers":   len(uniqueAcknowledgers),
	}

	acknowledgementStats := fiber.Map{
		"TotalAcknowledgements": len(acknowledgementsMade),
		"ActiveAcks":            activeAcks,
		"RefutedAcks":           refutedAcks,
		"UniqueExpressions":     len(uniqueExpressionsAcked),
		"UniqueCreators":        len(uniqueCreatorsAcked),
	}

	// Sort expressions by timestamp to get most recent
	recentExpressions := expressions
	if len(recentExpressions) > 5 {
		recentExpressions = recentExpressions[:5]
	}

	data := fiber.Map{
		"Title":                "Dashboard - Proof of Peacemaking",
		"User":                 fiber.Map{"Email": user.Email, "Address": user.Address},
		"ExpressionStats":      expressionStats,
		"AcknowledgementStats": acknowledgementStats,
		"RecentExpressions":    recentExpressions,
	}

	log.Printf("[DASHBOARD] Data being passed to template: %+v", data)
	return c.Render("dashboard", data, "")
}

func (h *DashboardHandler) GetExpressions(c *fiber.Ctx) error {
	log.Printf("[DASHBOARD] Starting expressions handler")

	// Get user identifier from context (set by auth middleware)
	userIdentifier, ok := c.Locals("userAddress").(string)
	if !ok {
		log.Printf("[DASHBOARD] Error: User identifier not found in context")
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user data",
		})
	}

	// Get user's expressions
	expressions, err := h.expressionService.ListByUser(c.Context(), user.ID.Hex())
	if err != nil {
		log.Printf("[DASHBOARD] Error fetching expressions: %v", err)
		expressions = []*domain.Expression{} // Use empty slice instead of failing
	}

	data := fiber.Map{
		"Title":       "My Expressions - Proof of Peacemaking",
		"User":        fiber.Map{"Email": user.Email, "Address": user.Address},
		"Expressions": expressions,
	}

	return c.Render("dashboard_expressions", data, "")
}

func (h *DashboardHandler) GetAcknowledgements(c *fiber.Ctx) error {
	log.Printf("[DASHBOARD] Starting acknowledgements handler")

	// Get user identifier from context (set by auth middleware)
	userIdentifier, ok := c.Locals("userAddress").(string)
	if !ok {
		log.Printf("[DASHBOARD] Error: User identifier not found in context")
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user data",
		})
	}

	// Get acknowledgements made by user
	acknowledgements, err := h.acknowledgementService.ListByUser(c.Context(), user.ID.Hex())
	if err != nil {
		log.Printf("[DASHBOARD] Error fetching acknowledgements: %v", err)
		acknowledgements = []*domain.Acknowledgement{} // Use empty slice instead of failing
	}

	// Collect all expression IDs
	expressionIDs := make([]string, 0, len(acknowledgements))
	for _, ack := range acknowledgements {
		expressionIDs = append(expressionIDs, ack.ExpressionID)
	}

	// Get all expressions in one query
	expressions, err := h.expressionService.GetMultiple(c.Context(), expressionIDs)
	if err != nil {
		log.Printf("[DASHBOARD] Error fetching expressions: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch expressions",
		})
	}

	data := fiber.Map{
		"Title":            "My Acknowledgements - Proof of Peacemaking",
		"User":             fiber.Map{"Email": user.Email, "Address": user.Address},
		"Acknowledgements": acknowledgements,
		"Expressions":      expressions,
	}

	return c.Render("dashboard_acknowledgements", data, "")
}

func (h *DashboardHandler) GetDashboardStats(c *fiber.Ctx) error {
	userIdentifier := c.Locals("userAddress").(string)
	if userIdentifier == "" {
		return c.Redirect("/")
	}

	var user *domain.User
	var err error

	if strings.Contains(userIdentifier, "@") {
		user, err = h.userService.GetUserByEmail(c.Context(), userIdentifier)
	} else {
		user, err = h.userService.GetUserByAddress(c.Context(), userIdentifier)
	}

	if err != nil {
		return c.Render("error", fiber.Map{
			"Error": "Failed to get user data",
		})
	}

	// Get user's expressions
	expressions, err := h.expressionService.ListByUser(c.Context(), user.ID.Hex())
	if err != nil {
		return c.Render("error", fiber.Map{
			"Error": "Failed to get expressions",
		})
	}

	// Get user's acknowledgments
	acknowledgments, err := h.acknowledgementService.ListByUser(c.Context(), user.ID.Hex())
	if err != nil {
		return c.Render("error", fiber.Map{
			"Error": "Failed to get acknowledgments",
		})
	}

	// Count acknowledgments by status
	ackStats := make(map[string]int)
	uniqueExpressions := make(map[string]bool)
	uniqueCreators := make(map[string]bool)

	for _, ack := range acknowledgments {
		// Count by status
		ackStats[string(ack.Status)]++

		// Track unique expressions
		uniqueExpressions[ack.ExpressionID] = true

		// Get expression to track creator
		expr, err := h.expressionService.Get(c.Context(), ack.ExpressionID)
		if err == nil && expr != nil {
			uniqueCreators[expr.CreatorAddress] = true
		}
	}

	data := fiber.Map{
		"Title":             "Dashboard",
		"User":              fiber.Map{"Email": user.Email, "Address": user.Address},
		"Expressions":       expressions,
		"TotalAcks":         len(acknowledgments),
		"ActiveAcks":        ackStats[string(domain.AcknowledgementStatusActive)],
		"RefutedAcks":       ackStats[string(domain.AcknowledgementStatusRefuted)],
		"UniqueExpressions": len(uniqueExpressions),
		"UniqueCreators":    len(uniqueCreators),
	}

	return c.Render("dashboard", data)
}
