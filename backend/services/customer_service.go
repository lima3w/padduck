package services

import (
	"context"
	"fmt"
	"strings"

	"padduck/models"
	"padduck/repository"
)

type CustomerService struct {
	repo *repository.Repository
}

func NewCustomerService(repo *repository.Repository) *CustomerService {
	return &CustomerService{repo: repo}
}

func (s *CustomerService) CreateCustomer(ctx context.Context, name, description, email, phone, notes string) (*models.Customer, error) {
	if name == "" {
		return nil, fmt.Errorf("customer name is required")
	}
	return s.repo.CreateCustomer(ctx, name, description, email, phone, notes)
}

func (s *CustomerService) GetCustomer(ctx context.Context, id int64) (*models.Customer, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid customer ID")
	}
	return s.repo.GetCustomerByID(ctx, id)
}

func (s *CustomerService) ListCustomers(ctx context.Context) ([]*models.Customer, error) {
	return s.repo.ListAllCustomers(ctx)
}

func (s *CustomerService) UpdateCustomer(ctx context.Context, id int64, name, description, email, phone, notes string) (*models.Customer, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid customer ID")
	}
	if name == "" {
		return nil, fmt.Errorf("customer name is required")
	}
	return s.repo.UpdateCustomer(ctx, id, name, description, email, phone, notes)
}

func (s *CustomerService) DeleteCustomer(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid customer ID")
	}
	return s.repo.DeleteCustomer(ctx, id)
}

func (s *CustomerService) ListCustomerAssociations(ctx context.Context, customerID int64) ([]*models.CustomerAssociation, error) {
	return s.repo.ListCustomerAssociations(ctx, customerID)
}

func (s *CustomerService) CreateCustomerAssociation(ctx context.Context, req *repository.CustomerAssociationParams) (*models.CustomerAssociation, error) {
	if req.CustomerID <= 0 || req.ObjectID <= 0 || strings.TrimSpace(req.ObjectType) == "" {
		return nil, fmt.Errorf("customer, object type, and object ID are required")
	}
	req.Relationship = defaultString(req.Relationship, "owner")
	return s.repo.CreateCustomerAssociation(ctx, req)
}

func (s *CustomerService) DeleteCustomerAssociation(ctx context.Context, id int64) error {
	return s.repo.DeleteCustomerAssociation(ctx, id)
}

func (s *CustomerService) ListCustomersPaginated(ctx context.Context, page, limit int) ([]*models.Customer, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 25
	}
	offset := (page - 1) * limit
	return s.repo.ListCustomersPaginated(ctx, limit, offset)
}
