package uuidstorage

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/klyakssa/test-repo-url/internal/config"
	"github.com/klyakssa/test-repo-url/internal/db/postgres"
	"github.com/klyakssa/test-repo-url/internal/filestorage"
)

type UUIDStorage struct {
	mu  sync.RWMutex
	bd  map[string]string
	fs  *filestorage.FileStorage
	cfg *config.Config
}

func New(cfg *config.Config) *UUIDStorage {
	fs, err := filestorage.New(cfg.File.Path)
	if err != nil {
		panic(err)
	}
	uuidstorage := &UUIDStorage{
		bd:  make(map[string]string),
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

func (s *UUIDStorage) Shorten(ctx context.Context, url string) (string, error) {
	done := make(chan struct{})
	var result string

	go func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		shurl := uuid.NewString()
		s.bd[shurl] = url
		result = fmt.Sprintf("%s/%s", s.cfg.WebConfig.BaseURL, shurl)

		close(done)
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-done:
		return result, nil
	}
}

func (s *UUIDStorage) Unshorten(ctx context.Context, uuid string) (string, error) {
	done := make(chan struct{})
	var result string

	go func() {
		s.mu.RLock()
		defer s.mu.RUnlock()
		result = s.bd[uuid]
		close(done)
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-done:
		return result, nil
	}
}

func (s *UUIDStorage) save(data map[string]string) *UUIDStorage {
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
