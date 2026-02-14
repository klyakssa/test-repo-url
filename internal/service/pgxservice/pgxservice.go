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

func (s *NewPgxService) Ping(ctx context.Context) error {
	return s.repo.Ping(ctx)
}
