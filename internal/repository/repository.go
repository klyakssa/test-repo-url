package repository

import (
	"sync"

	"github.com/google/uuid"
)

type Repository interface {
	Shorten(string) (string, error)
	Unshorten(string) (string, error)
}

var bd = make(map[string]string)
var mu sync.RWMutex

func Shorten(url string) (string, error) {
	mu.Lock()
	defer mu.Unlock()
	shurl := uuid.New().String()
	bd[shurl] = url
	return shurl, nil
}

func Unshorten(shurl string) (string, error) {
	mu.RLock()
	defer mu.RUnlock()
	return bd[shurl], nil
}
