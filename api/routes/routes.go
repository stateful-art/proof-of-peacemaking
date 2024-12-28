package routes

import (
	"proofofpeacemaking/internal/handlers"
	"proofofpeacemaking/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, h *handlers.Handlers) {
	// Home page
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{
			"Title": "Proof of Peacemaking",
		})
	})

	// Public routes
	app.Post("/auth/nonce", h.Auth.GenerateNonce)
	app.Post("/auth/verify", h.Auth.VerifySignature)

	// Protected routes
	api := app.Group("/api", middleware.AuthMiddleware())

	// Notification routes
	notifications := api.Group("/notifications")
	notifications.Get("/", h.Notification.GetUserNotifications)
	notifications.Put("/:id/read", h.Notification.MarkAsRead)

	// Expression routes
	expressions := api.Group("/expressions")
	expressions.Post("/", h.Expression.Create)
	expressions.Get("/", h.Expression.List)
	expressions.Get("/:id", h.Expression.Get)

	// Acknowledgement routes
	acknowledgements := api.Group("/acknowledgements")
	acknowledgements.Post("/", h.Acknowledgement.Create)
	acknowledgements.Get("/expression/:id", h.Acknowledgement.ListByExpression)

	// ProofNFT routes
	proofs := api.Group("/proofs")
	proofs.Post("/request", h.ProofNFT.RequestProof)
	proofs.Put("/approve/:id", h.ProofNFT.ApproveProof)
	proofs.Get("/user", h.ProofNFT.ListUserProofs)
}
