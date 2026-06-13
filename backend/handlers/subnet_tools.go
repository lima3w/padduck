package handlers

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"padduck/services"
)

// SplitSubnetRequest is the body for POST /api/v1/admin/subnets/:id/split
type SplitSubnetRequest struct {
	NewPrefixLen int `json:"new_prefix_len"`
}

// MergeSubnetsRequest is the body for POST /api/v1/admin/subnets/merge
type MergeSubnetsRequest struct {
	SubnetIDs []int64 `json:"subnet_ids"`
}

// ResizeSubnetRequest is the body for POST /api/v1/admin/subnets/:id/resize
type ResizeSubnetRequest struct {
	NewPrefix string `json:"new_prefix"`
}

// SplitSubnet handles POST /api/v1/admin/subnets/:id/split
func (h *Handler) SplitSubnet(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid subnet ID")
	}

	req := new(SplitSubnetRequest)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if req.NewPrefixLen <= 0 {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "new_prefix_len is required")
	}

	children, err := h.service.SplitSubnet(c.Context(), int64(id), req.NewPrefixLen)
	if err != nil {
		var blockedErr *services.SplitBlockedError
		if errors.As(err, &blockedErr) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error":        blockedErr.Error(),
				"blocking_ips": blockedErr.BlockingIPs,
			})
		}
		reqLogger(c).Error("split subnet error", "id", id, "error", err)
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}

	uid, uname := auditUserFromCtx(c)
	sid := int64(id)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "subnet.split",
		ResourceType: "subnet", ResourceID: &sid,
		NewValues: map[string]interface{}{
			"new_prefix_len": req.NewPrefixLen,
			"child_count":    len(children),
		},
	})

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"subnets": children})
}

// MergeSubnets handles POST /api/v1/admin/subnets/merge
func (h *Handler) MergeSubnets(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}

	req := new(MergeSubnetsRequest)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if len(req.SubnetIDs) < 2 {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "at least 2 subnet_ids required")
	}

	parent, err := h.service.MergeSubnets(c.Context(), req.SubnetIDs)
	if err != nil {
		reqLogger(c).Error("merge subnets error", "error", err)
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}

	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "subnet.merge",
		ResourceType: "subnet", ResourceID: &parent.ID,
		NewValues: map[string]interface{}{
			"merged_subnet_ids": req.SubnetIDs,
			"result_cidr":       fmt.Sprintf("%s/%d", parent.NetworkAddress, parent.PrefixLength),
		},
	})

	return c.Status(fiber.StatusCreated).JSON(parent)
}

// ResizeSubnet handles POST /api/v1/admin/subnets/:id/resize
func (h *Handler) ResizeSubnet(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid subnet ID")
	}
	if err := h.permCheck(c, services.PermV2SubnetWrite, services.ResourceScope{Type: "subnet", ID: int64(id)}); err != nil {
		return nil
	}

	req := new(ResizeSubnetRequest)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if req.NewPrefix == "" {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "new_prefix is required")
	}

	subnet, err := h.service.ResizeSubnet(c.Context(), int64(id), req.NewPrefix)
	if err != nil {
		var conflictErr *services.SubnetResizeConflictError
		if errors.As(err, &conflictErr) {
			resp := fiber.Map{"error": conflictErr.Error()}
			if len(conflictErr.ConflictingIPs) > 0 {
				resp["conflicting_ips"] = conflictErr.ConflictingIPs
			}
			if len(conflictErr.ConflictingSubnets) > 0 {
				resp["conflicting_subnets"] = conflictErr.ConflictingSubnets
			}
			return c.Status(fiber.StatusConflict).JSON(resp)
		}
		reqLogger(c).Error("resize subnet error", "id", id, "error", err)
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}

	uid, uname := auditUserFromCtx(c)
	sid := int64(id)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "subnet.resize",
		ResourceType: "subnet", ResourceID: &sid,
		NewValues: map[string]interface{}{"new_prefix": req.NewPrefix},
	})

	return c.JSON(subnet)
}
