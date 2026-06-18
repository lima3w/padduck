package services

import (
	"context"
	"fmt"
	"strings"

	"padduck/models"
	"padduck/repository"
)

func (s *Service) CreateCustomer(ctx context.Context, name, description, email, phone, notes string) (*models.Customer, error) {
	if name == "" {
		return nil, fmt.Errorf("customer name is required")
	}
	return s.repository.CreateCustomer(ctx, name, description, email, phone, notes)
}

func (s *Service) GetCustomer(ctx context.Context, id int64) (*models.Customer, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid customer ID")
	}
	return s.repository.GetCustomerByID(ctx, id)
}

func (s *Service) ListCustomers(ctx context.Context) ([]*models.Customer, error) {
	return s.repository.ListAllCustomers(ctx)
}

func (s *Service) UpdateCustomer(ctx context.Context, id int64, name, description, email, phone, notes string) (*models.Customer, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid customer ID")
	}
	if name == "" {
		return nil, fmt.Errorf("customer name is required")
	}
	return s.repository.UpdateCustomer(ctx, id, name, description, email, phone, notes)
}

func (s *Service) DeleteCustomer(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid customer ID")
	}
	return s.repository.DeleteCustomer(ctx, id)
}

func (s *Service) ListCustomerAssociations(ctx context.Context, customerID int64) ([]*models.CustomerAssociation, error) {
	return s.repository.ListCustomerAssociations(ctx, customerID)
}

func (s *Service) CreateCustomerAssociation(ctx context.Context, req *repository.CustomerAssociationParams) (*models.CustomerAssociation, error) {
	if req.CustomerID <= 0 || req.ObjectID <= 0 || strings.TrimSpace(req.ObjectType) == "" {
		return nil, fmt.Errorf("customer, object type, and object ID are required")
	}
	req.Relationship = defaultString(req.Relationship, "owner")
	return s.repository.CreateCustomerAssociation(ctx, req)
}

func (s *Service) DeleteCustomerAssociation(ctx context.Context, id int64) error {
	return s.repository.DeleteCustomerAssociation(ctx, id)
}
