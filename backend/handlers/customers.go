package handlers

import (
	"github.com/gofiber/fiber/v2"
	"ipam-next/models"
	"ipam-next/services"
)

type CreateCustomerRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Notes       string `json:"notes"`
}

type UpdateCustomerRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Notes       string `json:"notes"`
}

func (h *Handler) ListCustomers(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2CustomerList); err != nil {
		return nil
	}
	customers, err := h.service.ListCustomers(c.Context())
	if err != nil {
		reqLogger(c).Error("error listing customers", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	if customers == nil {
		customers = make([]*models.Customer, 0)
	}
	return c.JSON(customers)
}

func (h *Handler) GetCustomer(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2CustomerRead); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid customer ID"})
	}
	customer, err := h.service.GetCustomer(c.Context(), int64(id))
	if err != nil {
		reqLogger(c).Error("error getting customer", "id", id, "error", err)
		return respondCustomerASError(c, err, "customer")
	}
	return c.JSON(customer)
}

func (h *Handler) CreateCustomer(c *fiber.Ctx) error {
	req := new(CreateCustomerRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := h.permCheck(c, services.PermV2CustomerWrite); err != nil {
		return nil
	}
	customer, err := h.service.CreateCustomer(c.Context(), req.Name, req.Description, req.Email, req.Phone, req.Notes)
	if err != nil {
		reqLogger(c).Error("error creating customer", "error", err)
		return respondCustomerASError(c, err, "customer")
	}
	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "customer_created",
		ResourceType: "customer", ResourceID: &customer.ID, ResourceName: customer.Name,
		NewValues: map[string]string{"name": customer.Name, "email": customer.Email},
	})
	return c.Status(fiber.StatusCreated).JSON(customer)
}

func (h *Handler) UpdateCustomer(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid customer ID"})
	}
	if err := h.permCheck(c, services.PermV2CustomerWrite, services.ResourceScope{Type: "customer", ID: int64(id)}); err != nil {
		return nil
	}
	req := new(UpdateCustomerRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	customer, err := h.service.UpdateCustomer(c.Context(), int64(id), req.Name, req.Description, req.Email, req.Phone, req.Notes)
	if err != nil {
		reqLogger(c).Error("error updating customer", "id", id, "error", err)
		return respondCustomerASError(c, err, "customer")
	}
	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "customer_updated",
		ResourceType: "customer", ResourceID: &customer.ID, ResourceName: customer.Name,
		NewValues: map[string]string{"name": req.Name, "email": req.Email, "description": req.Description},
	})
	return c.JSON(customer)
}

func (h *Handler) DeleteCustomer(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid customer ID"})
	}
	if err := h.permCheck(c, services.PermV2CustomerDelete, services.ResourceScope{Type: "customer", ID: int64(id)}); err != nil {
		return nil
	}
	if err := h.service.DeleteCustomer(c.Context(), int64(id)); err != nil {
		reqLogger(c).Error("error deleting customer", "id", id, "error", err)
		return respondCustomerASError(c, err, "customer")
	}
	uid, uname := auditUserFromCtx(c)
	cid := int64(id)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "customer_deleted",
		ResourceType: "customer", ResourceID: &cid,
	})
	return c.SendStatus(fiber.StatusNoContent)
}
