package main

import (
	"fmt"
	"goSSR/auth"
	"goSSR/database"
	"goSSR/routes"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/session"
	postgresStorage "github.com/gofiber/storage/postgres/v3"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {

	// Load environment variables from
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize storage config
	storage := postgresStorage.New(postgresStorage.Config{
		ConnectionURI: os.Getenv("STORAGE_DB_URL"),
		Table:         os.Getenv("STORAGE_DB_TABLE"),
		SSLMode:       os.Getenv("STORAGE_DB_SSL_MODE"),
		Reset:         false,
		GCInterval:    10 * time.Second,
	})

	// Close the storage when the program terminates
	defer storage.Close()

	// Initialize Google OAuth config
	auth.InitializeOAuthConfig()

	// Set up session store
	store := session.New(session.Config{
		Expiration: 20 * time.Minute,
		Storage:    storage,
	})

	// Initialize database
	dsn := os.Getenv("DATABASE_URL")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	// migrate the schemas
	db.AutoMigrate(database.GetModels()...)

	// template engine
	engine := html.New("./views", ".html")

	app := fiber.New(fiber.Config{
		Views:        engine,
		ErrorHandler: CustomErrorHandler,
	})

	//static files
	app.Static("/public", "./public")

	// Favicon
	app.Use(favicon.New(favicon.Config{
		File: "./public/favicon.ico",
	}))

	app.Use(func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			// Use the custom error handler
			return CustomErrorHandler(c, fmt.Errorf("session error: %v", err))
		}
		c.Locals("session", sess)
		return c.Next()
	})

	// Security Headers
	app.Use(func(c *fiber.Ctx) error {
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		return c.Next()
	})

	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			if userID := c.Locals("userID"); userID != nil {
				return fmt.Sprintf("%s:%v", c.IP(), userID)
			}
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "Rate limit exceeded. Please try again later.",
				"retry_after": c.GetRespHeader("X-RateLimit-Reset"),
			})
		},
		SkipFailedRequests:     false,
		SkipSuccessfulRequests: false,
		LimiterMiddleware:      limiter.SlidingWindow{},
	}))

	// set up routes
	routes.Setup(app, db)

	log.Fatal(app.Listen(":3000"))
}
