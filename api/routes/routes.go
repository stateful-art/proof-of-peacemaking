package routes

import (
	"log"
	"proofofpeacemaking/internal/handlers"
	"proofofpeacemaking/internal/middleware"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func SetupRoutes(app *fiber.App, h *handlers.Handlers) {
	// Add CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// Add error handling middleware
	app.Use(func(c *fiber.Ctx) error {
		err := c.Next()
		if err != nil {
			log.Printf("[ERROR] Path: %s, Error: %v", c.Path(), err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return nil
	})

	// Add request logging middleware
	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()
		path := c.Path()
		method := c.Method()
		log.Printf("[REQUEST] %s %s", method, path)
		if method == "GET" {
			log.Printf("[QUERY] %s", c.Context().QueryArgs().String())
		} else if method == "POST" {
			log.Printf("[BODY] %s", string(c.Body()))
		}

		err := c.Next()

		duration := time.Since(start)
		status := c.Response().StatusCode()
		log.Printf("[RESPONSE] %s %s - Status: %d - Duration: %v", method, path, status, duration)
		if err != nil {
			log.Printf("[ERROR] %s %s - Error: %v", method, path, err)
		}

		return err
	})

	// Create middleware using auth service from handlers
	authMiddleware := middleware.NewAuthMiddleware(h.Auth.GetAuthService())

	// Home page (public)
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{
			"Title": "Proof of Peacemaking",
		}, "")
	})

	// Public auth routes
	app.Get("/auth/nonce", h.Auth.GenerateNonce)
	app.Get("/auth/session", h.Auth.GetSession)
	app.Post("/auth/verify", h.Auth.VerifySignature)
	app.Post("/auth/register", h.Auth.Register)
	app.Post("/auth/logout", h.Auth.Logout)

	// Protected routes

	// Feed page (protected)
	app.Get("/feed", authMiddleware.Authenticate(), h.Feed.GetFeed)

	// Dashboard page (protected)
	app.Get("/dashboard", authMiddleware.Authenticate(), h.Dashboard.GetDashboard)

	api := app.Group("/api", authMiddleware.Authenticate())
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
