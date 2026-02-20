package uuidservice

import (
	"context"

	"github.com/klyakssa/test-repo-url/internal/repository"
)

type UUIDService struct {
	repo repository.Repository
}

func New(repo repository.Repository) *UUIDService {
	return &UUIDService{
		repo: repo,
	}
}

func (s *UUIDService) Shorten(url string, ctx context.Context) (string, error) {
	shurl, err := s.repo.Shorten(url, ctx)
	if err != nil {
		return "", err
	}
	return shurl, nil
}

func (s *UUIDService) Unshorten(uuid string, ctx context.Context) (string, error) {
	url, err := s.repo.Unshorten(uuid, ctx)
	if err != nil {
		return "", err
	}
	return url, nil
}

func (s *UUIDService) Close() error {
	return s.repo.Close()
}

func (s *UUIDService) PingContext(ctx context.Context) error {
	return s.repo.PingContext(ctx)
}
