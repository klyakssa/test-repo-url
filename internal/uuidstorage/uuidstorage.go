package uuidstorage

import (
	"context"
	"fmt"
	"sync"

	"github.com/gofiber/fiber/v2/log"
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
		log.Error(err)
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
		log.Error(err)
		panic(err)
	}
	return uuidstorage.save(data)
}

func (s *UUIDStorage) Shorten(url string, ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	shurl := uuid.NewString()
	s.bd[shurl] = url
	return fmt.Sprintf("%s/%s", s.cfg.WebConfig.BaseUrl, shurl), nil
}

func (s *UUIDStorage) Unshorten(uuid string, ctx context.Context) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.bd[uuid], nil
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
		log.Error(err)
		return err
	}
	return s.fs.Close()
}

func (s *UUIDStorage) PingContext(ctx context.Context) error {
	return postgres.ErrConnection
}
