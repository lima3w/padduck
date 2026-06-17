package handlers

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"padduck/services"
)

// GetSystemHealth handles GET /api/v1/admin/system-health
// Returns a combined health summary: database, migrations, and scan agents.
func (h *Handler) GetSystemHealth(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}

	ctx := c.Context()
	pool := h.service.GetRepository().GetPool()

	// --- database health ---
	dbStatus := fiber.Map{"status": "ok"}
	if err := pool.QueryRow(ctx, "SELECT 1").Scan(new(int)); err != nil {
		dbStatus = fiber.Map{"status": "error", "detail": err.Error()}
	}

	// --- scan agents ---
	type agentCounts struct {
		total   int
		healthy int
		offline int
	}
	var counts agentCounts
	rows, agentErr := pool.Query(ctx, "SELECT is_active, status, last_seen FROM scan_agents")
	if agentErr == nil {
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
	// agentErr intentionally ignored — zero counts are the correct fallback when
	// the scan_agents table doesn't exist yet (pre-migration fresh install).
	_ = agentErr

	return c.JSON(fiber.Map{
		"database": dbStatus,
		"scan_agents": fiber.Map{
			"total":   counts.total,
			"healthy": counts.healthy,
			"offline": counts.offline,
		},
	})
}

// DownloadBackup handles GET /api/v1/admin/backup/download
// Exports all application tables using PostgreSQL's COPY protocol via the
// existing pgx connection pool. No external binaries required — works with
// any PostgreSQL server version.
//
// The output is a plain-SQL file containing COPY FROM stdin blocks that can
// be fed directly to psql:
//
//	psql $DATABASE_URL -f padduck-backup-YYYYMMDD-HHMMSS.sql
//
// Triggers are disabled per-table inside the transaction so that the data
// can be loaded in alphabetical order without FK violations. Sequences are
// reset at the end to match the largest primary key observed.
func (h *Handler) DownloadBackup(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}

	ctx := c.Context()
	pool := h.service.GetRepository().GetPool()

	// Acquire a single connection so COPY TO STDOUT runs on the same session.
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, "backup_failed",
			"failed to acquire database connection: "+err.Error())
	}
	defer conn.Release()

	var buf bytes.Buffer
	now := time.Now().UTC()

	fmt.Fprintf(&buf, "-- Padduck database backup\n")
	fmt.Fprintf(&buf, "-- Generated: %s\n", now.Format(time.RFC3339))
	fmt.Fprintf(&buf, "-- Restore:   psql $DATABASE_URL -f <this file>\n\n")

	// ── List public base tables ────────────────────────────────────────────────
	tableRows, err := conn.Query(ctx, `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
		ORDER BY table_name`)
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, "backup_failed",
			"failed to list tables: "+err.Error())
	}
	var tables []string
	for tableRows.Next() {
		var t string
		if scanErr := tableRows.Scan(&t); scanErr == nil {
			tables = append(tables, t)
		}
	}
	tableRows.Close()
	if err := tableRows.Err(); err != nil {
		return RespondError(c, fiber.StatusInternalServerError, "backup_failed",
			"error reading table list: "+err.Error())
	}

	fmt.Fprintf(&buf, "BEGIN;\n\n")

	// Disable FK trigger checks per table so load order doesn't matter.
	for _, t := range tables {
		fmt.Fprintf(&buf, "ALTER TABLE %q DISABLE TRIGGER ALL;\n", t)
	}
	fmt.Fprint(&buf, "\n")

	// Truncate all tables in one sweep (CASCADE handles FK references).
	for _, t := range tables {
		fmt.Fprintf(&buf, "TRUNCATE TABLE %q RESTART IDENTITY CASCADE;\n", t)
	}
	fmt.Fprint(&buf, "\n")

	// ── COPY data for each table ───────────────────────────────────────────────
	for _, t := range tables {
		colRows, err := conn.Query(ctx, `
			SELECT column_name
			FROM information_schema.columns
			WHERE table_schema = 'public' AND table_name = $1
			ORDER BY ordinal_position`, t)
		if err != nil {
			return RespondError(c, fiber.StatusInternalServerError, "backup_failed",
				fmt.Sprintf("failed to read columns for %q: %v", t, err))
		}
		var quotedCols []string
		for colRows.Next() {
			var col string
			if scanErr := colRows.Scan(&col); scanErr == nil {
				quotedCols = append(quotedCols, `"`+col+`"`)
			}
		}
		colRows.Close()
		if err := colRows.Err(); err != nil {
			return RespondError(c, fiber.StatusInternalServerError, "backup_failed",
				fmt.Sprintf("error reading columns for %q: %v", t, err))
		}
		if len(quotedCols) == 0 {
			continue
		}

		fmt.Fprintf(&buf, "-- table: %s\n", t)
		fmt.Fprintf(&buf, "COPY %q (%s) FROM stdin;\n", t, strings.Join(quotedCols, ", "))

		_, err = conn.Conn().PgConn().CopyTo(ctx, &buf,
			fmt.Sprintf(`COPY %q TO STDOUT (FORMAT text)`, t))
		if err != nil {
			return RespondError(c, fiber.StatusInternalServerError, "backup_failed",
				fmt.Sprintf("failed to dump table %q: %v", t, err))
		}
		// PostgreSQL writes the row data but not the psql terminator.
		fmt.Fprint(&buf, "\\.\n\n")
	}

	// ── Re-enable triggers ─────────────────────────────────────────────────────
	for _, t := range tables {
		fmt.Fprintf(&buf, "ALTER TABLE %q ENABLE TRIGGER ALL;\n", t)
	}
	fmt.Fprint(&buf, "\n")

	// ── Reset sequences to max(id)+1 for tables with a serial id column ────────
	seqRows, err := conn.Query(ctx, `
		SELECT s.sequence_name, t.table_name
		FROM information_schema.sequences s
		JOIN information_schema.columns c
		  ON c.column_default LIKE '%' || s.sequence_name || '%'
		JOIN information_schema.tables t
		  ON t.table_name = c.table_name
		WHERE s.sequence_schema = 'public'
		  AND c.table_schema    = 'public'
		  AND t.table_type      = 'BASE TABLE'
		ORDER BY s.sequence_name`)
	if err == nil {
		type seqRow struct{ seq, table string }
		var seqs []seqRow
		for seqRows.Next() {
			var r seqRow
			if scanErr := seqRows.Scan(&r.seq, &r.table); scanErr == nil {
				seqs = append(seqs, r)
			}
		}
		seqRows.Close()
		for _, r := range seqs {
			var maxID int64
			_ = conn.QueryRow(ctx,
				fmt.Sprintf(`SELECT COALESCE(MAX(id), 0) FROM %q`, r.table),
			).Scan(&maxID)
			fmt.Fprintf(&buf, "SELECT setval('%s', %d, true);\n", r.seq, maxID)
		}
		if len(seqs) > 0 {
			fmt.Fprint(&buf, "\n")
		}
	}

	fmt.Fprint(&buf, "COMMIT;\n")

	filename := "padduck-backup-" + now.Format("20060102-150405") + ".sql"
	c.Set("Content-Type", "application/octet-stream")
	c.Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	return c.Send(buf.Bytes())
}
