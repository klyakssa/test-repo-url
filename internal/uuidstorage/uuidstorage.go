package uuidstorage

import (
	"sync"

	"github.com/google/uuid"
)

type UUIDStorage struct {
	mu sync.RWMutex
	bd map[string]string
}

func New() *UUIDStorage {
	return &UUIDStorage{
		bd: make(map[string]string),
		mu: sync.RWMutex{},
	}
}

func (s *UUIDStorage) Shorten(url string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	shurl := uuid.NewString()
	s.bd[shurl] = url
	return shurl, nil
}

func (s *UUIDStorage) Unshorten(shurl string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.bd[shurl], nil
}

func (s *UUIDStorage) Load() map[string]string {
	return s.bd
}

func (s *UUIDStorage) Save(data map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.bd = data
	return nil
}
