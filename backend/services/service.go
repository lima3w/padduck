package services

import (
	"ipam-next/repository"
)

type Service struct {
	repository *repository.Repository
}

func NewService(repo *repository.Repository) *Service {
	return &Service{repository: repo}
}

// GetRepository returns the underlying repository
func (s *Service) GetRepository() *repository.Repository {
	return s.repository
}
