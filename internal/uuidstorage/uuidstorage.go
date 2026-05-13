package uuidstorage

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/klyakssa/test-repo-url/internal/config"
	"github.com/klyakssa/test-repo-url/internal/db/postgres"
	"github.com/klyakssa/test-repo-url/internal/filestorage"
	"github.com/klyakssa/test-repo-url/internal/model"
)

type UUIDStorage struct {
	mu  sync.RWMutex
	bd  map[uuid.UUID]string
	fs  *filestorage.FileStorage
	cfg *config.Config
}

func New(cfg *config.Config) *UUIDStorage {
	fs, err := filestorage.New(cfg.File.Path)
	if err != nil {
		panic(err)
	}
	uuidstorage := &UUIDStorage{
		bd:  make(map[uuid.UUID]string),
		mu:  sync.RWMutex{},
		fs:  fs,
		cfg: cfg,
	}
	data, err := fs.Load()
	if err != nil {
		panic(err)
	}
	return uuidstorage.save(data)
}

func (s *UUIDStorage) Shorten(ctx context.Context, url model.Storage) (string, error) {
	done := make(chan struct{})
	errChan := make(chan error, 1)
	var result string

	go func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		uid := uuid.New()

		s.bd[uid] = url.OriginalURL

		url.ShortURL = uid.String()

		// var err error
		// result, err = urls.JoinPath(s.cfg.WebConfig.BaseURL, url.ShortURL)
		// if err != nil {
		// 	errChan <- fmt.Errorf("join path: %w", err)
		// }
		result = fmt.Sprintf("%s/%s", s.cfg.WebConfig.BaseURL, url.ShortURL)

		close(done)
	}()

	select {
	case err := <-errChan:
		return "", err
	case <-ctx.Done():
		return "", ctx.Err()
	case <-done:
		return result, nil
	}
}

func (s *UUIDStorage) Unshorten(ctx context.Context, uid model.Storage) (string, error) {
	done := make(chan struct{})
	var result string

	go func() {
		s.mu.RLock()
		defer s.mu.RUnlock()
		result = s.bd[uuid.MustParse(uid.UUID)]
		close(done)
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-done:
		return result, nil
	}
}

func (s *UUIDStorage) GetUrlsByUserID(ctx context.Context, userID string) ([]model.UrlsStorage, error) {
	return nil, nil
}

func (s *UUIDStorage) DeleteUrlsByUserID(ctx context.Context, userID string, uuids []string) error {
	return nil
}

func (s *UUIDStorage) save(data map[uuid.UUID]string) *UUIDStorage {
	s.mu.Lock()
	defer s.mu.Unlock()
	if data != nil {
		s.bd = data
	}
	return s
}

func (s *UUIDStorage) Close() error {
	if err := s.fs.Save(s.bd); err != nil {
		return err
	}
	return s.fs.Close()
}

func (s *UUIDStorage) PingContext(ctx context.Context) error {
	return postgres.ErrConnection
}
