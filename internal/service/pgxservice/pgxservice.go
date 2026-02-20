package pgxservice

import (
	"context"

	"github.com/klyakssa/test-repo-url/internal/logger"
	"github.com/klyakssa/test-repo-url/internal/repository"
)

type NewPgxService struct {
	log  *logger.MyLogger
	repo repository.PostgresRepository
}

func New(log *logger.MyLogger, repo repository.PostgresRepository) *NewPgxService {
	return &NewPgxService{
		log:  log,
		repo: repo,
	}
}

func (s *NewPgxService) Shorten(url string) (string, error) {
	shurl, err := s.repo.Shorten(url)
	if err != nil {
		return "", err
	}
	return shurl, nil
}

func (s *NewPgxService) Unshorten(shurl string) (string, error) {
	url, err := s.repo.Unshorten(shurl)
	if err != nil {
		return "", err
	}
	return url, nil
}

func (s *NewPgxService) PingContext(ctx context.Context) error {
	return s.repo.PingContext(ctx)
}
