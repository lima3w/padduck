package handlers

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"padduck/services"
)

// generateSQLBackup produces a plain-SQL backup of all application tables using
// PostgreSQL's COPY protocol. This is extracted from the legacy DownloadBackup handler.
func (h *Handler) generateSQLBackup(c *fiber.Ctx) ([]byte, error) {
	ctx := c.Context()
	pool := h.service.GetRepository().GetPool()

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	var buf bytes.Buffer
	now := time.Now().UTC()

	fmt.Fprintf(&buf, "-- Padduck database backup\n")
	fmt.Fprintf(&buf, "-- Generated: %s\n", now.Format(time.RFC3339))
	fmt.Fprintf(&buf, "-- Restore:   psql $DATABASE_URL -f <this file>\n\n")

	tableRows, err := conn.Query(ctx, `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
		ORDER BY table_name`)
	if err != nil {
		return nil, fmt.Errorf("list tables: %w", err)
	}
	var tables []string
	for tableRows.Next() {
		var t string
		if scanErr := tableRows.Scan(&t); scanErr == nil {
			tables = append(tables, t)
		}
	}
	tableRows.Close()

	fmt.Fprintf(&buf, "BEGIN;\n\n")
	for _, t := range tables {
		fmt.Fprintf(&buf, "ALTER TABLE %q DISABLE TRIGGER ALL;\n", t)
	}
	fmt.Fprint(&buf, "\n")
	for _, t := range tables {
		fmt.Fprintf(&buf, "TRUNCATE TABLE %q RESTART IDENTITY CASCADE;\n", t)
	}
	fmt.Fprint(&buf, "\n")

	for _, t := range tables {
		colRows, err := conn.Query(ctx, `
			SELECT column_name
			FROM information_schema.columns
			WHERE table_schema = 'public' AND table_name = $1
			ORDER BY ordinal_position`, t)
		if err != nil {
			return nil, fmt.Errorf("read columns for %q: %w", t, err)
		}
		var quotedCols []string
		for colRows.Next() {
			var col string
			if scanErr := colRows.Scan(&col); scanErr == nil {
				quotedCols = append(quotedCols, `"`+col+`"`)
			}
		}
		colRows.Close()
		if len(quotedCols) == 0 {
			continue
		}
		fmt.Fprintf(&buf, "-- table: %s\n", t)
		fmt.Fprintf(&buf, "COPY %q (%s) FROM stdin;\n", t, strings.Join(quotedCols, ", "))
		_, err = conn.Conn().PgConn().CopyTo(ctx, &buf,
			fmt.Sprintf(`COPY %q TO STDOUT (FORMAT text)`, t))
		if err != nil {
			return nil, fmt.Errorf("dump table %q: %w", t, err)
		}
		fmt.Fprint(&buf, "\\.\n\n")
	}

	for _, t := range tables {
		fmt.Fprintf(&buf, "ALTER TABLE %q ENABLE TRIGGER ALL;\n", t)
	}
	fmt.Fprint(&buf, "\n")

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
	return buf.Bytes(), nil
}

// DownloadFullBackup handles GET /api/v1/admin/backups/download
// Produces a ZIP archive containing:
//   - db/padduck-backup.sql  — full PostgreSQL COPY dump
//   - config/settings.json  — all config key/value pairs
//   - files/**              — contents of ./data/ directory
//   - backup-manifest.json  — metadata
func (h *Handler) DownloadFullBackup(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}

	ctx := c.Context()
	pool := h.service.GetRepository().GetPool()
	now := time.Now().UTC()

	// ── Build ZIP in memory ────────────────────────────────────────────────────
	var zipBuf bytes.Buffer
	zw := zip.NewWriter(&zipBuf)

	// 1. Manifest
	manifest := map[string]string{
		"version":    "1",
		"created_at": now.Format(time.RFC3339),
		"app":        "padduck",
	}
	manifestBytes, _ := json.MarshalIndent(manifest, "", "  ")
	if mf, err := zw.Create("backup-manifest.json"); err == nil {
		_, _ = mf.Write(manifestBytes)
	}

	// 2. SQL database dump
	sqlBytes, err := h.generateSQLBackup(c)
	if err != nil {
		_ = zw.Close()
		return RespondError(c, fiber.StatusInternalServerError, "backup_failed", "database backup failed: "+err.Error())
	}
	if sf, err := zw.Create("db/padduck-backup.sql"); err == nil {
		_, _ = sf.Write(sqlBytes)
	}

	// 3. Config settings
	rows, err := pool.Query(ctx, `SELECT key, value FROM configs ORDER BY key`)
	if err == nil {
		configMap := make(map[string]string)
		for rows.Next() {
			var k, v string
			if scanErr := rows.Scan(&k, &v); scanErr == nil {
				configMap[k] = v
			}
		}
		rows.Close()
		configBytes, _ := json.MarshalIndent(configMap, "", "  ")
		if cf, err := zw.Create("config/settings.json"); err == nil {
			_, _ = cf.Write(configBytes)
		}
	}

	// 4. Data files (./data/)
	wd, _ := os.Getwd()
	dataDir := filepath.Join(wd, "data")
	if info, statErr := os.Stat(dataDir); statErr == nil && info.IsDir() {
		dataRoot, rootErr := os.OpenRoot(dataDir)
		if rootErr != nil {
			return RespondError(c, fiber.StatusInternalServerError, "backup_failed", "failed to open data directory: "+rootErr.Error())
		}
		defer dataRoot.Close()
		_ = filepath.Walk(dataDir, func(path string, fi os.FileInfo, walkErr error) error {
			if walkErr != nil || fi.IsDir() {
				return nil
			}
			relData, relErr := filepath.Rel(dataDir, path)
			if relErr != nil {
				return nil
			}
			// Use forward slashes in zip entry names
			entryName := "files/data/" + filepath.ToSlash(relData)
			if fw, createErr := zw.Create(entryName); createErr == nil {
				f, openErr := dataRoot.Open(relData)
				if openErr == nil {
					_, _ = io.Copy(fw, f)
					_ = f.Close()
				}
			}
			return nil
		})
	}

	if err := zw.Close(); err != nil {
		return RespondError(c, fiber.StatusInternalServerError, "backup_failed", "failed to finalize zip: "+err.Error())
	}

	filename := "padduck-backup-" + now.Format("20060102-150405") + ".zip"
	c.Set("Content-Type", "application/zip")
	c.Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	return c.Send(zipBuf.Bytes())
}

const maxRestoreFileSize = 500 * 1024 * 1024 // 500 MB

// RestoreFromBackup handles POST /api/v1/admin/backups/restore
// Accepts a multipart/form-data upload with a "file" field containing a .zip
// backup archive produced by DownloadFullBackup.
func (h *Handler) RestoreFromBackup(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}

	fh, err := c.FormFile("file")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, "bad_request", "file field is required")
	}
	if fh.Size > maxRestoreFileSize {
		return RespondError(c, fiber.StatusRequestEntityTooLarge, "file_too_large", "file too large (max 500 MB)")
	}

	f, err := fh.Open()
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, "restore_failed", "cannot open uploaded file")
	}
	defer f.Close()

	// Read into memory so zip.NewReader can seek
	data, err := io.ReadAll(f)
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, "restore_failed", "cannot read uploaded file")
	}

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, "bad_request", "not a valid zip archive")
	}

	// Validate manifest
	manifestFound := false
	for _, zf := range zr.File {
		if zf.Name == "backup-manifest.json" {
			rc, openErr := zf.Open()
			if openErr != nil {
				break
			}
			var manifest map[string]string
			if decodeErr := json.NewDecoder(rc).Decode(&manifest); decodeErr == nil {
				if manifest["version"] == "1" {
					manifestFound = true
				}
			}
			_ = rc.Close()
			break
		}
	}
	if !manifestFound {
		return RespondError(c, fiber.StatusBadRequest, "bad_request", "invalid backup archive: missing or incompatible manifest")
	}

	ctx := c.Context()
	pool := h.service.GetRepository().GetPool()

	// Restore each section
	for _, zf := range zr.File {
		rc, openErr := zf.Open()
		if openErr != nil {
			continue
		}
		content, readErr := io.ReadAll(rc)
		_ = rc.Close()
		if readErr != nil {
			continue
		}

		switch {
		case zf.Name == "db/padduck-backup.sql":
			// Execute the SQL dump using the low-level pgconn Exec
			conn, acquireErr := pool.Acquire(ctx)
			if acquireErr != nil {
				return RespondError(c, fiber.StatusInternalServerError, "restore_failed", "cannot acquire database connection")
			}
			_, execErr := conn.Exec(ctx, string(content))
			conn.Release()
			if execErr != nil {
				return RespondError(c, fiber.StatusInternalServerError, "restore_failed", "database restore failed: "+execErr.Error())
			}

		case zf.Name == "config/settings.json":
			var configMap map[string]string
			if jsonErr := json.Unmarshal(content, &configMap); jsonErr != nil {
				continue
			}
			for k, v := range configMap {
				_, _ = pool.Exec(ctx,
					`INSERT INTO configs (key, value) VALUES ($1, $2)
					 ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value`,
					k, v)
			}

		case strings.HasPrefix(zf.Name, "files/"):
			// Restore file to ./data/... stripping the "files/" prefix
			rel := strings.TrimPrefix(zf.Name, "files/")
			if rel == "" || strings.Contains(rel, "..") {
				continue // skip empty or path-traversal entries
			}
			wd, _ := os.Getwd()
			dest := filepath.Join(wd, filepath.FromSlash(rel))
			// Security: ensure dest is still under wd
			if !strings.HasPrefix(dest, wd+string(os.PathSeparator)) {
				continue
			}
			if mkErr := os.MkdirAll(filepath.Dir(dest), 0750); mkErr != nil {
				continue
			}
			if writeErr := os.WriteFile(dest, content, 0600); writeErr != nil {
				continue // log but continue
			}
		}
	}

	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "backup_restored",
		ResourceType: "backup",
		NewValues:    map[string]string{"source": fh.Filename},
	})

	return c.JSON(fiber.Map{"message": "Backup restored successfully"})
}
