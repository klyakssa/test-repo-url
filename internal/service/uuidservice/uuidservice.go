package uuidservice

import (
	"github.com/klyakssa/go-musthave-shortener-tpl/internal/repository"
	"github.com/klyakssa/go-musthave-shortener-tpl/internal/uuidstorage"
)

type UUIDService struct {
	repo repository.Repository
}

func New() *UUIDService {
	return &UUIDService{
		repo: uuidstorage.New(),
	}
}

func (s *UUIDService) Shorten(url string) (string, error) {
	shurl, err := s.repo.Shorten(url)
	if err != nil {
		return "", err
	}
	return shurl, nil
}

func (s *UUIDService) Unshorten(shurl string) (string, error) {
	url, err := s.repo.Unshorten(shurl)
	if err != nil {
		return "", err
	}
	return url, nil
}

func (s *UUIDService) Load() map[string]string {
	return s.repo.Load()
}

func (s *UUIDService) Save(data map[string]string) error {
	if data == nil {
		return nil
	}
	return s.repo.Save(data)
}
