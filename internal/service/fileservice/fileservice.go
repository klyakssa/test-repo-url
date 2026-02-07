package fileservice

import (
	"github.com/klyakssa/go-musthave-shortener-tpl/internal/config"
	"github.com/klyakssa/go-musthave-shortener-tpl/internal/filestorage"
	"github.com/klyakssa/go-musthave-shortener-tpl/internal/repository"
)

type FileService struct {
	repo repository.FileStorageRepository
}

func New(cfg *config.Config) *FileService {
	fst, err := filestorage.New(cfg.File.Path)
	if err != nil {
		panic(err)
	}
	return &FileService{repo: fst}
}

func (f *FileService) Save(data map[string]string) error {
	return f.repo.Save(data)
}

func (f *FileService) Load() (map[string]string, error) {
	return f.repo.Load()
}

func (f *FileService) Delete() error {
	return f.repo.Delete()
}

func (f *FileService) Close() error {
	return f.repo.Close()
}
