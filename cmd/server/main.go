package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"time"

	"proofofpeacemaking/api/routes"
	"proofofpeacemaking/internal/config"
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"
	"proofofpeacemaking/internal/core/services"
	"proofofpeacemaking/internal/core/storage"
	"proofofpeacemaking/internal/handlers"
	"proofofpeacemaking/internal/repositories/mongodb"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/mailgun/mailgun-go/v4"
	"go.mongodb.org/mongo-driver/mongo"
)

func initializeServices(db *mongo.Database, r2Storage storage.Storage, mailgunClient *mailgun.MailgunImpl) (
	ports.UserService,
	ports.AuthService,
	ports.ExpressionService,
	ports.AcknowledgementService,
	ports.ProofNFTService,
	ports.FeedService,
	ports.NewsletterService,
	ports.WebAuthnService,
	ports.SessionService,
	ports.StatisticsService,
	ports.SongService,
	ports.ConversationService,
	ports.NotificationService,
	error,
) {
	// Initialize repositories
	userRepo := mongodb.NewUserRepository(db)
	expressionRepo := mongodb.NewExpressionRepository(db)
	acknowledgementRepo := mongodb.NewAcknowledgementRepository(db)
	proofNFTRepo := mongodb.NewProofNFTRepository(db)
	sessionRepo := mongodb.NewSessionRepository(db)
	statsRepo := mongodb.NewStatisticsRepository(db)
	passkeyRepo := mongodb.NewPasskeyRepository(db)
	songRepo := mongodb.NewSongRepository(db)
	conversationRepo := mongodb.NewConversationRepository(db)
	notificationRepo := mongodb.NewNotificationRepository(db)

	// Initialize LiveKit client
	livekitHost := os.Getenv("LIVEKIT_HOST")
	livekitAPIKey := os.Getenv("LIVEKIT_API_KEY")
	livekitAPISecret := os.Getenv("LIVEKIT_API_SECRET")
	livekitClient := lksdk.NewRoomServiceClient(livekitHost, livekitAPIKey, livekitAPISecret)

	// Initialize services
	userService := services.NewUserService(userRepo)
	authService := services.NewAuthService(userService, sessionRepo)
	expressionService := services.NewExpressionService(expressionRepo, acknowledgementRepo, r2Storage)
	acknowledgementService := services.NewAcknowledgementService(acknowledgementRepo)
	proofNFTService := services.NewProofNFTService(userRepo, proofNFTRepo)
	feedService := services.NewFeedService(expressionService, userService, acknowledgementService)
	newsletterService := services.NewNewsletterService(mailgunClient)
	webAuthnService, err := services.NewWebAuthnService(passkeyRepo, userRepo)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	sessionService := services.NewSessionService(sessionRepo)
	statsService := services.NewStatisticsService(statsRepo, userRepo, expressionRepo)
	songService := services.NewSongService(songRepo)
	notificationService := services.NewNotificationService(notificationRepo)
	conversationService := services.NewConversationService(conversationRepo, livekitClient, notificationService)

	// Ensure indexes
	if err := songRepo.EnsureIndexes(); err != nil {
		log.Printf("Failed to ensure song indexes: %v", err)
	}

	return userService, authService, expressionService, acknowledgementService, proofNFTService, feedService, newsletterService, webAuthnService, sessionService, statsService, songService, conversationService, notificationService, nil
}

func getProjectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "../..")
}

func setupHandlers(app *fiber.App) *handlers.Handlers {
	// Setup MongoDB connection
	db := mongodb.Connect()
	mailgunClient := getMailgunClient()

	// Initialize R2 storage
	expressionsR2Storage, err := initExpressionsR2Storage()
	if err != nil {
		log.Fatalf("Failed to initialize expressions R2 storage: %v", err)
	}

	// Initialize services
	userService, authService, expressionService, acknowledgementService, proofNFTService, feedService, newsletterService, webAuthnService, sessionService, statsService, songService, conversationService, notificationService, err := initializeServices(db, expressionsR2Storage, mailgunClient)
	if err != nil {
		log.Fatalf("Failed to initialize services: %v", err)
	}

	// Initialize handlers
	handlers := handlers.NewHandlers(
		userService,
		authService,
		expressionService,
		acknowledgementService,
		proofNFTService,
		feedService,
		statsService,
		webAuthnService,
		sessionService,
		newsletterService,
		songService,
		conversationService,
		notificationService,
	)

	// Setup routes with user service for feed handler
	routes.SetupRoutes(app, handlers)

	return handlers
}

func initTemplateEngine() *html.Engine {
	engine := html.New("./web/templates", ".html")
	engine.Reload(true)

	// Add template functions
	engine.AddFunc("trimAddress", func(address string) string {
		if len(address) <= 10 {
			return address
		}
		return address[:6] + "..." + address[len(address)-4:]
	})

	engine.AddFunc("formatDate", func(date time.Time) string {
		return date.Format("Jan 02, 2006")
	})

	// Load all templates
	if err := engine.Load(); err != nil {
		log.Printf("Warning: Failed to load templates: %v", err)
	}

	return engine
}

func initExpressionsR2Storage() (*storage.R2Storage, error) {
	expressionsConfig, err := config.GetR2Config("EXPRESSIONS")
	if err != nil {
		log.Fatalf("Failed to get expressions R2 config: %v", err)
	}
	expressionsR2Storage, err := storage.NewR2Storage(
		expressionsConfig.S3AccessKeyID,
		expressionsConfig.S3SecretKey,
		expressionsConfig.AccountID,
		expressionsConfig.Bucket,
	)
	if err != nil {
		log.Fatalf("Failed to initialize expressions R2 storage: %v", err)
	}

	return expressionsR2Storage, nil
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	// Load country data
	if err := domain.LoadCountries("web/static/data/countries.json"); err != nil {
		log.Fatalf("Failed to load country data: %v", err)
	}

	// Get project root directory
	projectRoot := getProjectRoot()

	loadEnvironment(projectRoot)

	// Setup template engine
	engine := initTemplateEngine()
	engine.Reload(true)

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

	// Setup handlers and routes
	setupHandlers(app)

	// Start server in a goroutine
	go func() {
		port := getPort()
		if err := app.Listen(port); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Use graceful shutdown
	gracefulShutdown(app)
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

// add a graceful shutdown and use it in the main function
func gracefulShutdown(app *fiber.App) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Fatalf("Failed to shutdown server: %v", err)
	}
}
