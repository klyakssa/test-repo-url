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

func (s *UUIDService) Shorten(ctx context.Context, url string) (string, error) {
	return s.repo.Shorten(ctx, url)
}

func (s *UUIDService) Unshorten(ctx context.Context, uuid string) (string, error) {
	return s.repo.Unshorten(ctx, uuid)
}

func (s *UUIDService) Close() error {
	return s.repo.Close()
}

func (s *UUIDService) PingContext(ctx context.Context) error {
	return s.repo.PingContext(ctx)
}
