package handlers

import (
	"bytes"
	"context"
	"io"

	"github.com/gofiber/fiber/v2"
	"padduck/services"
)

// ─────────────────────────────────────────────────────────────────────────────
// CSV import (#225 #226 #227)
// ─────────────────────────────────────────────────────────────────────────────

const maxImportFileSize = 5 * 1024 * 1024 // 5 MB

// ImportSubnetsCSV handles POST /api/v1/admin/import/subnets
// Accepts multipart/form-data with a "file" field containing a CSV.
func (h *Handler) ImportSubnetsCSV(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}

	fh, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "file field is required"})
	}
	if fh.Size > maxImportFileSize {
		return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{"error": "file too large (max 5 MB)"})
	}

	f, err := fh.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "cannot open uploaded file"})
	}
	defer f.Close()

	if c.QueryBool("dry_run") {
		result, err := h.service.Import.DryRunSubnetsCSV(c.Context(), f)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(result)
	}

	if c.QueryBool("async") {
		data, err := io.ReadAll(f)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "cannot read uploaded file"})
		}
		job := h.service.Jobs.Enqueue("import", "Import subnets CSV", fiber.Map{"source": "subnets_csv"}, 2, func(ctx context.Context, reporter *services.JobReporter) (interface{}, error) {
			reporter.Progress(0, 1, "importing subnets")
			result, err := h.service.Import.ImportSubnetsCSV(ctx, bytes.NewReader(data))
			reporter.Progress(1, 1, "import complete")
			return result, err
		})
		return c.Status(fiber.StatusAccepted).JSON(job)
	}

	result, err := h.service.Import.ImportSubnetsCSV(c.Context(), f)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}

// ImportIPsCSV handles POST /api/v1/admin/import/ips
// Accepts multipart/form-data with a "file" field containing a CSV.
func (h *Handler) ImportIPsCSV(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}

	fh, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "file field is required"})
	}
	if fh.Size > maxImportFileSize {
		return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{"error": "file too large (max 5 MB)"})
	}

	f, err := fh.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "cannot open uploaded file"})
	}
	defer f.Close()

	if c.QueryBool("dry_run") {
		result, err := h.service.Import.DryRunIPsCSV(c.Context(), f)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(result)
	}

	if c.QueryBool("async") {
		data, err := io.ReadAll(f)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "cannot read uploaded file"})
		}
		job := h.service.Jobs.Enqueue("import", "Import IPs CSV", fiber.Map{"source": "ips_csv"}, 2, func(ctx context.Context, reporter *services.JobReporter) (interface{}, error) {
			reporter.Progress(0, 1, "importing IP addresses")
			result, err := h.service.Import.ImportIPsCSV(ctx, bytes.NewReader(data))
			reporter.Progress(1, 1, "import complete")
			return result, err
		})
		return c.Status(fiber.StatusAccepted).JSON(job)
	}

	result, err := h.service.Import.ImportIPsCSV(c.Context(), f)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}

// ImportFromPHPIpam handles POST /api/v1/admin/import/phpipam
// Query param: kind=subnets|ips
// Accepts multipart/form-data with a "file" field.
func (h *Handler) ImportFromPHPIpam(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}

	kind := c.Query("kind")
	if kind != "subnets" && kind != "ips" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "kind query param must be \"subnets\" or \"ips\""})
	}

	fh, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "file field is required"})
	}
	if fh.Size > maxImportFileSize {
		return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{"error": "file too large (max 5 MB)"})
	}

	f, err := fh.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "cannot open uploaded file"})
	}
	defer f.Close()

	if c.QueryBool("dry_run") {
		var result interface{}
		var dryErr error
		switch kind {
		case "subnets":
			result, dryErr = h.service.Import.DryRunPHPIpamSubnetsCSV(c.Context(), f)
		case "ips":
			result, dryErr = h.service.Import.DryRunPHPIpamIPsCSV(c.Context(), f)
		}
		if dryErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": dryErr.Error()})
		}
		return c.JSON(result)
	}

	if c.QueryBool("async") {
		data, err := io.ReadAll(f)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "cannot read uploaded file"})
		}
		job := h.service.Jobs.Enqueue("import", "Import phpIPAM "+kind, fiber.Map{"source": "phpipam", "kind": kind}, 2, func(ctx context.Context, reporter *services.JobReporter) (interface{}, error) {
			reporter.Progress(0, 1, "importing phpIPAM "+kind)
			result, err := h.service.Import.ImportFromPHPIpam(ctx, bytes.NewReader(data), kind)
			reporter.Progress(1, 1, "import complete")
			return result, err
		})
		return c.Status(fiber.StatusAccepted).JSON(job)
	}

	result, err := h.service.Import.ImportFromPHPIpam(c.Context(), f, kind)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}

// ─────────────────────────────────────────────────────────────────────────────
// Full data export (#228)
// ─────────────────────────────────────────────────────────────────────────────

// ExportFullData handles GET /api/v1/admin/export/full
// Query param: format=csv|json (default csv)
func (h *Handler) ExportFullData(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminRead); err != nil {
		return nil
	}

	format := c.Query("format", "csv")

	if c.QueryBool("async") {
		job := h.service.Jobs.Enqueue("export", "Export full data", fiber.Map{"format": format}, 2, func(ctx context.Context, reporter *services.JobReporter) (interface{}, error) {
			reporter.Progress(0, 1, "building export")
			data, filename, contentType, err := h.service.Import.ExportFullData(ctx, format)
			reporter.Progress(1, 1, "export complete")
			return fiber.Map{"filename": filename, "content_type": contentType, "bytes": len(data)}, err
		})
		return c.Status(fiber.StatusAccepted).JSON(job)
	}

	data, filename, contentType, err := h.service.Import.ExportFullData(c.Context(), format)
	if err != nil {
		reqLogger(c).Error("export full data failed", "format", format, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	return c.Send(data)
}

// ExportV2MigrationBundle handles GET /api/v1/admin/export/v2-migration-bundle.
func (h *Handler) ExportV2MigrationBundle(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminRead); err != nil {
		return nil
	}

	if c.QueryBool("async") {
		job := h.service.Jobs.Enqueue("export", "Export v2 migration bundle", fiber.Map{"target": "v2.0"}, 2, func(ctx context.Context, reporter *services.JobReporter) (interface{}, error) {
			reporter.Progress(0, 1, "building v2 migration bundle")
			data, filename, contentType, err := h.service.Import.ExportV2MigrationBundle(ctx)
			reporter.Progress(1, 1, "migration bundle export complete")
			return fiber.Map{"filename": filename, "content_type": contentType, "bytes": len(data)}, err
		})
		return c.Status(fiber.StatusAccepted).JSON(job)
	}

	data, filename, contentType, err := h.service.Import.ExportV2MigrationBundle(c.Context())
	if err != nil {
		reqLogger(c).Error("export v2 migration bundle failed", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	return c.Send(data)
}
