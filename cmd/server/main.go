package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/stateful-art/proof-of-peacemaking/api/routes"
	"github.com/stateful-art/proof-of-peacemaking/internal/handlers"
	"github.com/stateful-art/proof-of-peacemaking/internal/repositories/mongodb"
	"github.com/stateful-art/proof-of-peacemaking/internal/services"
)

func setupHandlers(app *fiber.App) *handlers.Handlers {
	// Setup MongoDB connection
	db := mongodb.Connect()

	// Setup repositories
	notificationRepo := mongodb.NewNotificationRepository(db)
	userRepo := mongodb.NewUserRepository(db)

	// Setup services
	notificationService := services.NewNotificationService(notificationRepo, userRepo)
	authService := services.NewAuthService(userRepo)

	// Create handlers
	h := handlers.NewHandlers(notificationService, authService)

	// Setup routes
	routes.SetupRoutes(app, h)

	return h
}

func main() {
	app := fiber.New()

	// Setup handlers and routes
	setupHandlers(app)

	log.Fatal(app.Listen(":3000"))
}
