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
	"github.com/mailgun/mailgun-go/v4"
	"go.mongodb.org/mongo-driver/mongo"
)

func initServices(db *mongo.Database, mailgunClient *mailgun.MailgunImpl) (
	ports.NotificationService,
	ports.AuthService,
	ports.ExpressionService,
	ports.AcknowledgementService,
	ports.ProofNFTService,
	ports.FeedService,
	ports.UserService,
	ports.NewsletterService,
) {
	// Initialize repositories
	userRepo := mongodb.NewUserRepository(db)
	sessionRepo := mongodb.NewSessionRepository(db)
	expressionRepo := mongodb.NewExpressionRepository(db)
	acknowledgementRepo := mongodb.NewAcknowledgementRepository(db)
	notificationRepo := mongodb.NewNotificationRepository(db)
	proofNFTRepo := mongodb.NewProofNFTRepository(db)

	// Initialize services
	userService := services.NewUserService(userRepo)
	authService := services.NewAuthService(userService, sessionRepo)
	expressionService := services.NewExpressionService(expressionRepo, acknowledgementRepo)
	acknowledgementService := services.NewAcknowledgementService(acknowledgementRepo)
	notificationService := services.NewNotificationService(notificationRepo, userRepo)
	proofNFTService := services.NewProofNFTService(userRepo, proofNFTRepo)
	feedService := services.NewFeedService(expressionService, userService, acknowledgementService)
	newsletterService := services.NewNewsletterService(mailgunClient)
	return notificationService, authService, expressionService, acknowledgementService, proofNFTService, feedService, userService, newsletterService
}

func getProjectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "../..")
}

func setupHandlers(app *fiber.App) *handlers.Handlers {
	// Setup MongoDB connection
	db := mongodb.Connect()
	mailgunClient := getMailgunClient()
	// Initialize services
	notificationService, authService, expressionService, acknowledgementService, proofNFTService, feedService, userService, newsletterService := initServices(db, mailgunClient)

	// Create handlers
	h := handlers.NewHandlers(
		notificationService,
		authService,
		expressionService,
		acknowledgementService,
		proofNFTService,
		feedService,
		userService,
		newsletterService,
	)

	// Setup routes with user service for feed handler
	routes.SetupRoutes(app, h)

	return h
}

func initTemplateEngine() *html.Engine {
	engine := html.New("./web/templates", ".html")

	// Add template functions
	engine.AddFunc("trimAddress", func(address string) string {
		if len(address) <= 10 {
			return address
		}
		return address[:6] + "..." + address[len(address)-4:]
	})

	return engine
}

func main() {
	// Get project root directory
	projectRoot := getProjectRoot()

	loadEnvironment(projectRoot)

	// Setup template engine
	engine := initTemplateEngine()
	engine.Reload(true) // Enable this for development
	engine.Debug(true)  // Enable debug mode for development

	// Add template functions
	engine.AddFunc("formatDate", func(date time.Time) string {
		return date.Format("Jan 02, 2006")
	})

	// Create Fiber app
	app := fiber.New(fiber.Config{
		Views: engine,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			log.Printf("Error handling request: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Setup static files
	app.Static("/static", filepath.Join(projectRoot, "web/static"))
	app.Static("/uploads", filepath.Join(projectRoot, "uploads"))

	// Create uploads directories if they don't exist
	uploadsPath := filepath.Join(projectRoot, "uploads")
	for _, dir := range []string{"images", "audio", "video"} {
		if err := os.MkdirAll(filepath.Join(uploadsPath, dir), 0755); err != nil {
			log.Printf("Warning: Failed to create uploads directory %s: %v", dir, err)
		}
	}

	// Setup handlers and routes
	setupHandlers(app)

	startServer(app)
}

func loadEnvironment(projectRoot string) {
	if _, exists := os.LookupEnv("RAILWAY_ENVIRONMENT"); !exists {
		// if err := godotenv.Load(); err != nil {
		// 	log.Fatal("error loading .env file:", err)
		// }

		// Load environment variables
		if err := godotenv.Load(filepath.Join(projectRoot, ".env")); err != nil {
			log.Printf("Warning: .env file not found")
		}
	}
}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = ":3000"
	} else {
		port = ":" + port
	}
	return port
}

func startServer(app *fiber.App) {
	port := getPort()
	log.Fatal(app.Listen(port))
}

func getMailgunClient() *mailgun.MailgunImpl {
	var domain string = os.Getenv("EMAIL_SENDER_DOMAIN")
	var key string = os.Getenv("MAILGUN_APIKEY")
	return mailgun.NewMailgun(domain, key)
}
