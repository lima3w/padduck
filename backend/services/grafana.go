package services

import (
	"context"

	"padduck/repository"
)

func (s *Service) GrafanaSubnetUtilisation(ctx context.Context) ([]repository.GrafanaSubnetRow, error) {
	return s.repository.GrafanaGetSubnetUtilisation(ctx)
}

func (s *Service) GrafanaIPCountsByStatus(ctx context.Context) ([]repository.GrafanaIPStatusRow, error) {
	return s.repository.GrafanaGetIPCountsByStatus(ctx)
}

func (s *Service) GrafanaNetworkSummary(ctx context.Context) ([]repository.GrafanaSectionRow, error) {
	return s.repository.GrafanaGetSectionSummary(ctx)
}
