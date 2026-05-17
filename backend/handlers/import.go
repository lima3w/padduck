package handlers

import (
	"github.com/gofiber/fiber/v2"
	"ipam-next/services"
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

	data, filename, contentType, err := h.service.Import.ExportFullData(c.Context(), format)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	return c.Send(data)
}
