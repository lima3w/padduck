package handlers

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	csvexport "padduck/internal/export"
	"padduck/services"
)

// ExportNetworksCSV handles GET /api/v1/admin/reports/export/networks
func (h *Handler) ExportNetworksCSV(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminRead); err != nil {
		return nil
	}

	sections, _, err := h.service.ListNetworksPaginated(c.Context(), 1, 10000)
	if err != nil {
		reqLogger(c).Error("export sections CSV failed", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	_ = w.Write([]string{"id", "name", "description"})
	for _, s := range sections {
		_ = w.Write([]string{
			fmt.Sprintf("%d", s.ID),
			csvexport.EscapeCSVCell(s.Name),
			csvexport.EscapeCSVCell(s.Description),
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		reqLogger(c).Error("export sections CSV flush failed", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	filename := "sections-" + time.Now().Format("20060102") + ".csv"
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	return c.Send(buf.Bytes())
}

// ExportDevicesCSV handles GET /api/v1/admin/reports/export/devices
func (h *Handler) ExportDevicesCSV(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminRead); err != nil {
		return nil
	}

	devices, err := h.service.ListAllDevices(c.Context())
	if err != nil {
		reqLogger(c).Error("export devices CSV failed", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	_ = w.Write([]string{"id", "hostname", "description", "vendor", "model", "os_version", "is_online", "ip_count"})
	for _, d := range devices {
		description := ""
		if d.Description != nil {
			description = *d.Description
		}
		vendor := ""
		if d.Vendor != nil {
			vendor = *d.Vendor
		}
		model := ""
		if d.Model != nil {
			model = *d.Model
		}
		osVersion := ""
		if d.OSVersion != nil {
			osVersion = *d.OSVersion
		}
		_ = w.Write([]string{
			fmt.Sprintf("%d", d.ID),
			csvexport.EscapeCSVCell(d.Hostname),
			csvexport.EscapeCSVCell(description),
			csvexport.EscapeCSVCell(vendor),
			csvexport.EscapeCSVCell(model),
			csvexport.EscapeCSVCell(osVersion),
			strconv.FormatBool(d.IsOnline),
			fmt.Sprintf("%d", d.IPCount),
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		reqLogger(c).Error("export devices CSV flush failed", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	filename := "devices-" + time.Now().Format("20060102") + ".csv"
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	return c.Send(buf.Bytes())
}

// ExportVLANsCSV handles GET /api/v1/admin/reports/export/vlans
func (h *Handler) ExportVLANsCSV(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminRead); err != nil {
		return nil
	}

	vlans, err := h.service.ListVLANs(c.Context())
	if err != nil {
		reqLogger(c).Error("export VLANs CSV failed", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	_ = w.Write([]string{"id", "vlan_id", "name", "description"})
	for _, v := range vlans {
		_ = w.Write([]string{
			fmt.Sprintf("%d", v.ID),
			fmt.Sprintf("%d", v.VlanID),
			csvexport.EscapeCSVCell(v.Name),
			csvexport.EscapeCSVCell(v.Description),
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		reqLogger(c).Error("export VLANs CSV flush failed", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	filename := "vlans-" + time.Now().Format("20060102") + ".csv"
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	return c.Send(buf.Bytes())
}

// ExportVRFsCSV handles GET /api/v1/admin/reports/export/vrfs
func (h *Handler) ExportVRFsCSV(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminRead); err != nil {
		return nil
	}

	vrfs, err := h.service.ListVRFs(c.Context())
	if err != nil {
		reqLogger(c).Error("export VRFs CSV failed", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	_ = w.Write([]string{"id", "name", "route_distinguisher", "description"})
	for _, v := range vrfs {
		_ = w.Write([]string{
			fmt.Sprintf("%d", v.ID),
			csvexport.EscapeCSVCell(v.Name),
			csvexport.EscapeCSVCell(v.RouteDistinguisher),
			csvexport.EscapeCSVCell(v.Description),
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		reqLogger(c).Error("export VRFs CSV flush failed", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	filename := "vrfs-" + time.Now().Format("20060102") + ".csv"
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	return c.Send(buf.Bytes())
}
