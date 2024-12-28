package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/stateful-art/proof-of-peacemaking/internal/handlers"
	"github.com/stateful-art/proof-of-peacemaking/internal/middleware"
)

func SetupRoutes(app *fiber.App, h *handlers.Handlers) {
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
