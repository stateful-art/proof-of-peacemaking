package main

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
)

func getProjectRoot() string {
	// Get the absolute path of the current file
	_, filename, _, _ := runtime.Caller(0)
	// Go up two directories from cmd/server/main.go to reach project root
	return filepath.Join(filepath.Dir(filename), "../..")
}

func main() {
	// Get project root directory
	projectRoot := getProjectRoot()
	log.Printf("Project root: %s", projectRoot)

	// Load environment variables from project root
	if err := godotenv.Load(filepath.Join(projectRoot, ".env")); err != nil {
		log.Printf("Warning: .env file not found")
	}

	// Setup template engine with absolute path
	engine := html.New(filepath.Join(projectRoot, "web/templates"), ".html")
	engine.Reload(true) // Enable this for development

	// Create new Fiber app
	app := fiber.New(fiber.Config{
		Views: engine,
		// Add custom error handler
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Middleware
	app.Use(logger.New())

	// Serve static files from project root
	app.Static("/static", filepath.Join(projectRoot, "web/static"))

	// Routes
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{
			"Title": "Proof of Peacemaking",
		})
	})

	app.Get("/dashboard", func(c *fiber.Ctx) error {
		return c.Render("dashboard", fiber.Map{
			"Title": "Dashboard",
		})
	})

	// API routes
	api := app.Group("/api")

	api.Post("/expressions", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Create expression endpoint",
		})
	})

	api.Post("/acknowledgments", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Create acknowledgment endpoint",
		})
	})

	api.Get("/expressions/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Get expression details endpoint",
			"id":      c.Params("id"),
		})
	})

	// Add this route before the other routes
	app.Get("/favicon.ico", func(c *fiber.Ctx) error {
		return c.SendFile("web/static/favicon.ico")
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Server starting on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatal("Error starting server: ", err)
	}
}
