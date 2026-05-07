
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
	"strings"
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
			if err := writePasswordFile(password); err != nil {
				slog.Warn("could not write admin password file; set ADMIN_PASSWORD explicitly", "error", err)
			} else {
				log.Printf("  Generated admin password written to /run/ipam/admin-password (mode 0600).")
			}
		} else {
			log.Printf("  Admin password reset from ADMIN_PASSWORD env var.")
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
		if err := writePasswordFile(password); err != nil {
			slog.Warn("could not write admin password file; set ADMIN_PASSWORD explicitly", "error", err)
		} else {
			log.Printf("  Generated admin password written to /run/ipam/admin-password (mode 0600).")
		}
		log.Printf("  Set ADMIN_PASSWORD env var to override.")
		log.Printf("========================================")
	}

	return nil
}

// writePasswordFile writes the admin password to /run/ipam/admin-password with mode 0600.
func writePasswordFile(password string) error {
	const dir = "/run/ipam"
	const path = "/run/ipam/admin-password"
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}
	if err := os.WriteFile(path, []byte(password), 0600); err != nil {
		return fmt.Errorf("writing password file %s: %w", path, err)
	}
	return nil
}

// parseTrustedProxies splits a comma-separated list of CIDRs/IPs and returns it.
// When the input is empty the default ["127.0.0.1"] is returned.
func parseTrustedProxies(s string) []string {
	if s == "" {
		return []string{"127.0.0.1"}
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return []string{"127.0.0.1"}
	}
	return out
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
	trustedProxies := parseTrustedProxies(os.Getenv("TRUSTED_PROXIES"))
	app := fiber.New(fiber.Config{
		// Trust X-Real-IP only from known proxy addresses.
		EnableTrustedProxyCheck: true,
		TrustedProxies:          trustedProxies,
		ProxyHeader:             "X-Real-IP",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			msg := "internal server error"
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
				msg = e.Message
			}
			slog.Error("request error", "method", c.Method(), "path", c.Path(), "status", code, "error", err.Error())
			return c.Status(code).JSON(fiber.Map{"error": msg, "code": "REQUEST_ERROR"})
		},
	})

	// Panic recovery — returns 500 instead of crashing
	app.Use(recover.New(recover.Config{
		EnableStackTrace: cfg.Environment != "production",
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
			slog.Error("health check database ping failed", "error", err)
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status":   "degraded",
				"database": "disconnected",
				"error":    "database unavailable",
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
