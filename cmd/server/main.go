package main

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"proofofpeacemaking/api/routes"
	"proofofpeacemaking/internal/core/ports"
	"proofofpeacemaking/internal/core/services"
	"proofofpeacemaking/internal/handlers"
	"proofofpeacemaking/internal/repositories/mongodb"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
)

func initServices(db *mongo.Database) (
	ports.NotificationService,
	ports.AuthService,
	ports.ExpressionService,
	ports.AcknowledgementService,
	ports.ProofNFTService,
	ports.FeedService,
	ports.UserService,
) {
	// Initialize repositories
	userRepo := mongodb.NewUserRepository(db)
	expressionRepo := mongodb.NewExpressionRepository(db)
	acknowledgementRepo := mongodb.NewAcknowledgementRepository(db)
	sessionRepo := mongodb.NewSessionRepository(db)
	notificationRepo := mongodb.NewNotificationRepository(db)

	// Initialize services
	userService := services.NewUserService(userRepo)
	expressionService := services.NewExpressionService(expressionRepo)
	acknowledgementService := services.NewAcknowledgementService(acknowledgementRepo)
	authService := services.NewAuthService(userService, sessionRepo)
	notificationService := services.NewNotificationService(notificationRepo, userRepo)
	proofNFTService := services.NewProofNFTService(userRepo)
	feedService := services.NewFeedService(expressionService, userService)

	return notificationService, authService, expressionService, acknowledgementService, proofNFTService, feedService, userService
}

func getProjectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "../..")
}

func setupHandlers(app *fiber.App) *handlers.Handlers {
	// Setup MongoDB connection
	db := mongodb.Connect()

	// Initialize services
	notificationService, authService, expressionService, acknowledgementService, proofNFTService, feedService, userService := initServices(db)

	// Create handlers
	h := handlers.NewHandlers(
		notificationService,
		authService,
		expressionService,
		acknowledgementService,
		proofNFTService,
		feedService,
		userService,
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
	engine.Reload(true)     // Enable this for development
	engine.Layout("layout") // Set the default layout template
	engine.Debug(true)      // Enable debug mode for development

	// Add template functions
	engine.AddFunc("formatDate", func(date time.Time) string {
		return date.Format("Jan 02, 2006")
	})

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
