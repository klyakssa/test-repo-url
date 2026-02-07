package filestorage

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/klyakssa/go-musthave-shortener-tpl/internal/model"
)

type FileStorage struct {
	file *os.File
}

func New(path string) (*FileStorage, error) {
	file, err := os.OpenFile(filepath.Join(path), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	return &FileStorage{file: file}, nil
}

func (f *FileStorage) Close() error {
	return f.file.Close()
}

func (f *FileStorage) Save(data map[string]string) error {
	if len(data) == 0 {
		return nil
	}
	i := 1
	var stf []model.FileStorageData
	for k, v := range data {
		stf = append(stf, model.FileStorageData{Uuid: strconv.Itoa(i), S_URL: k, O_URL: v})
		i++
	}
	dt, err := json.Marshal(stf)
	if err != nil {
		return err
	}

	err = f.file.Truncate(0)
	if err != nil {
		return err
	}

	_, err = f.file.Seek(0, 0)
	if err != nil {
		return err
	}

	_, err = f.file.Write(dt)
	return err
}

func (f *FileStorage) Load() (map[string]string, error) {
	data, err := io.ReadAll(f.file)
	if err != nil {
		panic(err)
	}
	if len(data) == 0 {
		return nil, nil
	}
	var stf []model.FileStorageData
	err = json.Unmarshal(data, &stf)
	if err != nil {
		return nil, err
	}
	m := make(map[string]string)
	for _, v := range stf {
		m[v.S_URL] = v.O_URL
	}
	return m, nil
}

func (f *FileStorage) Delete() error {
	return os.Remove(f.file.Name())
}
