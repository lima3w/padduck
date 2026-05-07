
package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"ipam-next/config"
	"ipam-next/db"
	"ipam-next/handlers"
	"ipam-next/repository"
	"ipam-next/services"
)

func initLogging() {
	env := os.Getenv("ENVIRONMENT")
	var h slog.Handler
	if env == "production" {
		h = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		h = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})
	}
	// SetDefault also bridges the standard log package so existing log.Printf calls
	// will flow through the structured handler.
	slog.SetDefault(slog.New(h))
	log.SetFlags(0) // slog adds timestamps; suppress duplicate from log package
}

func initAdminPassword(ctx context.Context, svc *services.Service) error {
	password := os.Getenv("ADMIN_PASSWORD")
	forceReset := os.Getenv("RESET_ADMIN_PASSWORD") == "true"
	generated := password == ""

	if generated {
		b := make([]byte, 16)
		if _, err := rand.Read(b); err != nil {
			return fmt.Errorf("generating random password: %w", err)
		}
		password = base64.RawURLEncoding.EncodeToString(b)
	}

	if forceReset {
		if err := svc.ForceResetAdminPassword(ctx, password); err != nil {
			return err
		}
		log.Printf("========================================")
		if generated {
			log.Printf("  Admin password RESET to:  %s", password)
		} else {
			log.Printf("  Admin password reset from ADMIN_PASSWORD.")
		}
		log.Printf("  Unset RESET_ADMIN_PASSWORD to disable this on next boot.")
		log.Printf("========================================")
		return nil
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
	initLogging()

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

	// Start notification queue worker
	svc.Notification.StartWorker(ctx)

	// Start discovery scheduler
	svc.Discovery.StartScheduler(ctx)

	// Setup HTTP server with centralized error handler
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			slog.Error("request error", "method", c.Method(), "path", c.Path(), "status", code, "error", err.Error())
			return c.Status(code).JSON(fiber.Map{"error": err.Error(), "code": "REQUEST_ERROR"})
		},
	})

	// Panic recovery — returns 500 instead of crashing
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
			slog.Error("panic recovered", "path", c.Path(), "panic", fmt.Sprintf("%v", e))
		},
	}))

	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} ${method} ${path} ${latency} ${ip}\n",
	}))

	// Serve OpenAPI spec
	app.Get("/api/openapi.yaml", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/yaml")
		return c.SendFile("./docs/openapi.yaml")
	})

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
