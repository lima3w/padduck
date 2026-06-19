package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"padduck/models"
	"padduck/repository"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type OrganizationService struct {
	repo *repository.Repository
}

func NewOrganizationService(repo *repository.Repository) *OrganizationService {
	return &OrganizationService{repo: repo}
}

func (s *OrganizationService) Create(ctx context.Context, name, slug string) (*models.Organization, error) {
	name = strings.TrimSpace(name)
	slug = strings.TrimSpace(slug)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if slug == "" {
		return nil, fmt.Errorf("slug is required")
	}
	if !slugPattern.MatchString(slug) {
		return nil, fmt.Errorf("slug must be lowercase alphanumeric with hyphens (e.g. my-org)")
	}
	return s.repo.CreateOrganization(ctx, name, slug)
}

func (s *OrganizationService) Get(ctx context.Context, id int64) (*models.Organization, error) {
	return s.repo.GetOrganization(ctx, id)
}

func (s *OrganizationService) List(ctx context.Context) ([]*models.Organization, error) {
	orgs, err := s.repo.ListOrganizations(ctx)
	if err != nil {
		return nil, err
	}
	if orgs == nil {
		orgs = []*models.Organization{}
	}
	return orgs, nil
}

func (s *OrganizationService) Delete(ctx context.Context, id int64) error {
	return s.repo.DeleteOrganization(ctx, id)
}

// EnsureDefault creates the default organization and seeds existing users if no orgs exist.
func (s *OrganizationService) EnsureDefault(ctx context.Context) (int64, error) {
	return s.repo.EnsureDefaultOrganization(ctx)
}
