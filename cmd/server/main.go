package main

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"proofofpeacemaking/api/routes"
	"proofofpeacemaking/internal/core/services"
	"proofofpeacemaking/internal/handlers"
	"proofofpeacemaking/internal/repositories/mongodb"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
)

func getProjectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "../..")
}

func setupHandlers(app *fiber.App) *handlers.Handlers {
	// Setup MongoDB connection
	db := mongodb.Connect()

	// Setup repositories
	notificationRepo := mongodb.NewNotificationRepository(db)
	userRepo := mongodb.NewUserRepository(db)

	// Setup services
	notificationService := services.NewNotificationService(notificationRepo, userRepo)
	authService := services.NewAuthService(userRepo)
	expressionService := services.NewExpressionService(userRepo)
	acknowledgementService := services.NewAcknowledgementService(userRepo)
	proofNFTService := services.NewProofNFTService(userRepo)

	// Create handlers
	h := handlers.NewHandlers(
		notificationService,
		authService,
		expressionService,
		acknowledgementService,
		proofNFTService,
	)

	// Setup routes
	routes.SetupRoutes(app, h)

	return h
}

func main() {
	// Get project root directory
	projectRoot := getProjectRoot()

	// Load environment variables
	if err := godotenv.Load(filepath.Join(projectRoot, ".env")); err != nil {
		log.Printf("Warning: .env file not found")
	}

	// Setup template engine
	engine := html.New(filepath.Join(projectRoot, "web/templates"), ".html")
	engine.Reload(true) // Enable this for development

	// Create Fiber app
	app := fiber.New(fiber.Config{
		Views: engine,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Setup static files
	app.Static("/static", filepath.Join(projectRoot, "web/static"))

	// Setup handlers and routes
	setupHandlers(app)

	// Get port from environment variable
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(app.Listen(":" + port))
}
