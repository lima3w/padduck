package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"ipam-next/services"
)

// GetSystemHealth handles GET /api/v1/admin/system-health
// Returns a combined health summary: database, migrations, scan agents, and
// backup restore rehearsal documentation steps.
func (h *Handler) GetSystemHealth(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}

	ctx := c.Context()
	pool := h.service.GetRepository().GetPool()

	// --- database health ---
	dbStatus := fiber.Map{"status": "ok"}
	if err := pool.QueryRow(ctx, "SELECT 1").Scan(new(int)); err != nil {
		dbStatus = fiber.Map{"status": "error", "detail": err.Error()}
	}

	// --- migrations ---
	migrationsInfo := fiber.Map{"status": "unknown"}
	var migCount int
	err := pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM schema_migrations",
	).Scan(&migCount)
	if err == nil {
		migrationsInfo = fiber.Map{"status": "ok", "applied": migCount}
	}

	// --- scan agents ---
	type agentCounts struct {
		total   int
		healthy int
		offline int
	}
	var counts agentCounts
	rows, err := pool.Query(ctx, "SELECT is_active, status, last_seen FROM scan_agents")
	if err == nil {
		defer rows.Close()
		now := time.Now()
		for rows.Next() {
			var isActive bool
			var status string
			var lastSeen *time.Time
			if scanErr := rows.Scan(&isActive, &status, &lastSeen); scanErr != nil {
				continue
			}
			counts.total++
			effectiveStatus := status
			if lastSeen != nil && now.Sub(*lastSeen) > 5*time.Minute && effectiveStatus != "offline" {
				effectiveStatus = "offline"
			}
			switch effectiveStatus {
			case "offline":
				counts.offline++
			default:
				if isActive {
					counts.healthy++
				} else {
					counts.offline++
				}
			}
		}
	}

	// --- backup / restore rehearsal notes (static) ---
	backupNotes := []map[string]string{
		{"step": "1", "action": "Verify backup", "detail": "Confirm latest pg_dump or WAL archive is accessible and recent."},
		{"step": "2", "action": "Restore to staging", "detail": "Run restore to a staging instance and verify schema version matches production."},
		{"step": "3", "action": "Smoke test", "detail": "Run /api/v1/health and spot-check key tables (subnets, ip_addresses, users)."},
		{"step": "4", "action": "DNS validation", "detail": "If DNS integration is enabled, verify nameserver connectivity after restore."},
		{"step": "5", "action": "Cutover", "detail": "Update load balancer or DNS to point to the restored instance. Monitor logs for 15 minutes."},
	}

	return c.JSON(fiber.Map{
		"database":   dbStatus,
		"migrations": migrationsInfo,
		"scan_agents": fiber.Map{
			"total":   counts.total,
			"healthy": counts.healthy,
			"offline": counts.offline,
		},
		"backup_notes": backupNotes,
	})
}
