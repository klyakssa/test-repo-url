package pgxservice

import (
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

func (s *NewPgxService) Ping() error {
	return s.repo.Ping()
}
