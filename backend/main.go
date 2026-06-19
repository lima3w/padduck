package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"padduck/config"
	"padduck/db"
	"padduck/handlers"
	"padduck/repository"
	"padduck/services"
)

// logLevelFromEnv maps LOG_LEVEL to a slog level. The default is warn:
// warnings and errors only. Set LOG_LEVEL=info for request/operational
// logging or LOG_LEVEL=debug for full verbosity.
func logLevelFromEnv(raw string) (slog.Level, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "debug":
		return slog.LevelDebug, true
	case "info":
		return slog.LevelInfo, true
	case "", "warn", "warning":
		return slog.LevelWarn, true
	case "error":
		return slog.LevelError, true
	default:
		return slog.LevelWarn, false
	}
}

func initLogging() {
	rawLevel := os.Getenv("LOG_LEVEL")
	level, known := logLevelFromEnv(rawLevel)

	// ENVIRONMENT controls only the output format; LOG_LEVEL controls verbosity.
	env := os.Getenv("ENVIRONMENT")
	format := "text"
	var h slog.Handler
	if env == "production" {
		format = "json"
		h = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	} else {
		h = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	}
	// SetDefault also bridges the standard log package so existing log.Printf calls
	// will flow through the structured handler (at info level).
	slog.SetDefault(slog.New(h))
	log.SetFlags(0) // slog adds timestamps; suppress duplicate from log package

	if !known {
		slog.Warn("unrecognized LOG_LEVEL, falling back to warn", "value", rawLevel)
	}
	// Emitted at warn so the active configuration is visible under the default level.
	slog.Warn("logging configured", "level", level.String(), "format", format)
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
			if p, err := writePasswordFile(password); err != nil {
				slog.Warn("could not write admin password file; set ADMIN_PASSWORD explicitly", "error", err)
			} else {
				log.Printf("  Generated admin password written to %s (mode 0600).", p)
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
		if p, err := writePasswordFile(password); err != nil {
			slog.Warn("could not write admin password file; set ADMIN_PASSWORD explicitly", "error", err)
		} else {
			log.Printf("  Generated admin password written to %s (mode 0600).", p)
		}
		log.Printf("  Set ADMIN_PASSWORD env var to override.")
		log.Printf("========================================")
	}

	return nil
}

// writePasswordFile writes the admin password to data/admin-password relative to
// the working directory with mode 0600. Using a local data subdirectory avoids
// requiring write access to system paths like /run/ipam when the process runs as
// a non-root user, and the directory can be bind-mounted from the host.
func writePasswordFile(password string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}
	dir := filepath.Join(wd, "data")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("creating data directory %s: %w", dir, err)
	}
	path := filepath.Join(dir, "admin-password")
	if err := os.WriteFile(path, []byte(password), 0600); err != nil { // #nosec G703 -- path is os.Getwd()+fixed suffix, not user input.
		return "", fmt.Errorf("writing password file %s: %w", path, err)
	}
	return path, nil
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
	migrateDryRun := flag.Bool("migrate-dry-run", false, "Print pending migrations and their SQL without applying, then exit.")
	flag.Parse()

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

	// --migrate-dry-run: show pending migrations without applying, then exit.
	if *migrateDryRun {
		if _, err := database.DryRunMigrations("./migrations"); err != nil {
			log.Fatalf("migrate-dry-run: %v", err)
		}
		return
	}

	// Run migrations
	if err := database.RunMigrations("./migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Warn if a v1 API compat sunset date has been configured.
	if cfg.V1CompatSunset != "" {
		slog.Warn("v1 API compat mode: v1 routes will be retired after the configured sunset date",
			"v1_compat_sunset", cfg.V1CompatSunset,
			"action", "migrate consumers to /api/v2 before this date",
		)
	}

	// Setup application layers
	repo := repository.NewRepository(database.Pool())
	svc := services.NewService(repo, cfg.MFAEncryptionKey)
	handler := handlers.NewHandler(svc, svc.Ops, svc.Auth, cfg.Environment == "production")
	handler.StartTokenLimiterCleanup(ctx)

	// Initialize admin password on first boot
	if err := initAdminPassword(ctx, svc); err != nil {
		log.Fatalf("Failed to initialize admin password: %v", err)
	}

	// Ensure default organization exists (seeds existing users on upgrade)
	if _, err := svc.Ops.Organizations.EnsureDefault(ctx); err != nil {
		log.Fatalf("Failed to ensure default organization: %v", err)
	}

	// Seed platform admin from PLATFORM_ADMIN_EMAIL env var (idempotent)
	if email := strings.TrimSpace(os.Getenv("PLATFORM_ADMIN_EMAIL")); email != "" {
		if u, err := repo.GetUserByEmail(ctx, email); err == nil {
			if !u.IsPlatformAdmin {
				if err := repo.SetPlatformAdmin(ctx, u.ID, true); err != nil {
					log.Fatalf("Failed to set platform admin for %s: %v", email, err)
				}
				log.Printf("Platform admin enabled for user: %s", email)
			}
		} else {
			slog.Warn("PLATFORM_ADMIN_EMAIL set but user not found", "email", email)
		}
	}

	// Start notification queue worker
	svc.Auth.Notification.StartWorker(ctx)
	svc.Ops.Webhooks.StartWorker(ctx)

	// Start discovery scheduler
	svc.Ops.Discovery.StartScheduler(ctx)
	// Start retention pruner (#435)
	svc.Ops.Discovery.StartRetentionPruner(ctx)

	// Start reporting jobs (utilization snapshots + scheduled reports)
	svc.Ops.Reports.StartUtilizationSnapshotJob(ctx)
	svc.Ops.Reports.StartScheduledReportJob(ctx)

	// Start telemetry job (no-op unless opt-in is enabled in admin settings)
	svc.Ops.Telemetry.StartTelemetryJob(ctx)

	// Setup HTTP server with centralized error handler
	trustedProxies := parseTrustedProxies(os.Getenv("TRUSTED_PROXIES"))
	app := fiber.New(fiber.Config{
		// Trust X-Real-IP only from known proxy addresses.
		EnableTrustedProxyCheck: true,
		TrustedProxies:          trustedProxies,
		ProxyHeader:             "X-Real-IP",
		ReadTimeout:             30 * time.Second,
		WriteTimeout:            60 * time.Second,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			msg := "internal server error"
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
				msg = e.Message
			}
			rid, _ := c.Locals("requestID").(string)
			uid := c.Locals("userID")
			slog.Error("request error", "method", c.Method(), "path", c.Path(), "status", code, "error", err.Error(), "request_id", rid, "user_id", uid)
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
		Next: func(c *fiber.Ctx) bool {
			return c.Path() == "/health"
		},
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
		if err := app.Shutdown(); err != nil {
			log.Printf("Shutdown error: %v", err)
		}
	}()

	// Start server
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("Starting server on %s", addr)
	if err := app.Listen(addr); err != nil {
		log.Printf("Server error: %v", err)
	}
}
