package uuidservice

import (
	"context"

	"github.com/klyakssa/test-repo-url/internal/mapper"
	"github.com/klyakssa/test-repo-url/internal/model"
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

func (s *UUIDService) Shorten(ctx context.Context, url model.CreateShortURLInput) (string, error) {
	return s.repo.Shorten(ctx, model.Storage{
		OriginalURL: url.OriginalURL,
		UserID:      url.UserID,
	})
}

func (s *UUIDService) Unshorten(ctx context.Context, uuid model.GetShortURLInput) (string, error) {
	return s.repo.Unshorten(ctx, model.Storage{
		UUID:   uuid.UUID,
		UserID: uuid.UserID,
	})
}

func (s *UUIDService) GetUrlsByUserID(ctx context.Context, userID string) ([]model.UrlsResponse, error) {
	urls, err := s.repo.GetUrlsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return mapper.ToUrlsResponse(urls), nil
}

func (s *UUIDService) Close() error {
	return s.repo.Close()
}

func (s *UUIDService) PingContext(ctx context.Context) error {
	return s.repo.PingContext(ctx)
}
