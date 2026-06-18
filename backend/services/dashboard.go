package services

import (
	"context"

	"padduck/models"
)

// ListUsersPaginated returns a paginated list of users.
func (s *Service) ListUsersPaginated(ctx context.Context, page, limit int) ([]*models.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 25
	}
	offset := (page - 1) * limit
	return s.repository.ListUsersPaginated(ctx, limit, offset)
}

// ListLocationsPaginated returns a paginated list of locations.
func (s *Service) ListLocationsPaginated(ctx context.Context, page, limit int) ([]*models.Location, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 25
	}
	offset := (page - 1) * limit
	return s.repository.ListLocationsPaginated(ctx, limit, offset)
}

// ListCustomersPaginated returns a paginated list of customers.
func (s *Service) ListCustomersPaginated(ctx context.Context, page, limit int) ([]*models.Customer, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 25
	}
	offset := (page - 1) * limit
	return s.repository.ListCustomersPaginated(ctx, limit, offset)
}
