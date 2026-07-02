package repository

import (
	"context"

	"github.com/klyakssa/test-repo-url/internal/model"
)

// helpers is an interface for service functions
type helpers interface {
	Close() error
}

// Repository is an interface for repository functions
type Repository interface {
	Shorten(context.Context, model.Storage) (string, error)
	Unshorten(context.Context, model.Storage) (string, error)
	GetUrlsByUserID(context.Context, string) ([]model.UrlsStorage, error)
	DeleteUrlsByUserID(context.Context, string, []string) error
	PingContext(context.Context) error
	GetStats(context.Context) (model.GetStats, error)
	helpers
}

// UserService is an interface for service functions
type UserService interface {
	helpers
	Shorten(context.Context, model.CreateShortURLInput) (string, error)
	Unshorten(context.Context, model.GetShortURLInput) (string, error)
	GetUrlsByUserID(context.Context, string) ([]model.UrlsResponse, error)
	DeleteUrlsByUserID(context.Context, string, []string)
	PingContext(context.Context) error
	GetStats(context.Context) (model.GetStats, error)
}
