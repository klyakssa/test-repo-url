package uuidservice

import (
	"context"
	"time"

	"github.com/klyakssa/test-repo-url/internal/logger"
	"github.com/klyakssa/test-repo-url/internal/mapper"
	"github.com/klyakssa/test-repo-url/internal/model"
	"github.com/klyakssa/test-repo-url/internal/repository"
)

type UUIDService struct {
	repo       repository.Repository
	deleteChan chan model.DeleteTask
	logger     *logger.MyLogger
}

func New(ctx context.Context, repo repository.Repository, logger *logger.MyLogger) *UUIDService {
	service := &UUIDService{
		repo:       repo,
		deleteChan: make(chan model.DeleteTask, 100),
		logger:     logger,
	}
	go service.deleteWorker(ctx)
	return service
}

func (s *UUIDService) deleteWorker(ctx context.Context) {
	batch := []model.DeleteTask{}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case task := <-s.deleteChan:
			batch = append(batch, task)
			if len(batch) >= 50 {
				s.flush(ctx, batch)
				batch = nil
			}
		case <-ticker.C:
			if len(batch) > 0 {
				s.flush(ctx, batch)
				batch = nil
			}
		case <-ctx.Done():
			if len(batch) > 0 {
				flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				s.flush(flushCtx, batch)
			}
			return
		}
	}
}

func (s *UUIDService) flush(ctx context.Context, batch []model.DeleteTask) {
	userURLs := make(map[string][]string)

	for _, task := range batch {
		userURLs[task.UserID] = append(userURLs[task.UserID], task.UUIDs...)
	}

	for userID, uuids := range userURLs {
		err := s.repo.DeleteUrlsByUserID(ctx, userID, uuids)
		if err != nil {
			s.logger.Error(err)
		}
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

func (s *UUIDService) DeleteUrlsByUserID(ctx context.Context, userID string, uuids []string) {
	select {
	case s.deleteChan <- model.DeleteTask{
		UserID: userID,
		UUIDs:  uuids,
	}:
	default:
		s.repo.DeleteUrlsByUserID(ctx, userID, uuids)
	}
}

func (s *UUIDService) Close() error {
	return s.repo.Close()
}

func (s *UUIDService) PingContext(ctx context.Context) error {
	return s.repo.PingContext(ctx)
}
