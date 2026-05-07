
package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"ipam-next/config"
	"ipam-next/db"
	"ipam-next/handlers"
	"ipam-next/repository"
	"ipam-next/services"
)

func initAdminPassword(ctx context.Context, svc *services.Service) error {
	password := os.Getenv("ADMIN_PASSWORD")
	generated := false

	if password == "" {
		b := make([]byte, 16)
		if _, err := rand.Read(b); err != nil {
			return fmt.Errorf("generating random password: %w", err)
		}
		password = base64.RawURLEncoding.EncodeToString(b)
		generated = true
	}

	set, err := svc.InitAdminPassword(ctx, password)
	if err != nil {
		return err
	}

	if set && generated {
		log.Printf("========================================")
		log.Printf("  Admin password (first boot):  %s", password)
		log.Printf("  Set ADMIN_PASSWORD env var to override.")
		log.Printf("========================================")
	}

	return nil
}

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	database, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Run migrations
	if err := database.RunMigrations("./migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Setup application layers
	repo := repository.NewRepository(database.Pool())
	svc := services.NewService(repo, cfg.MFAEncryptionKey)
	handler := handlers.NewHandler(svc)

	// Initialize admin password on first boot
	if err := initAdminPassword(ctx, svc); err != nil {
		log.Fatalf("Failed to initialize admin password: %v", err)
	}

	// Setup HTTP server
	app := fiber.New()

	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} ${method} ${path} ${latency} ${ip}\n",
	}))

	// Health check endpoint with database verification
	app.Get("/health", func(c *fiber.Ctx) error {
		if err := repo.Ping(ctx); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status":   "degraded",
				"database": "disconnected",
				"error":    err.Error(),
			})
		}
		return c.JSON(fiber.Map{"status": "ok", "database": "connected"})
	})

	// Register all routes
	handler.RegisterRoutes(app)

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down gracefully...")
		app.Shutdown()
	}()

	// Start server
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("Starting server on %s", addr)
	if err := app.Listen(addr); err != nil {
		log.Printf("Server error: %v", err)
	}
}
