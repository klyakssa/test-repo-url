package repository

import (
	"context"

	"github.com/klyakssa/test-repo-url/internal/model"
)

type helpers interface {
	Close() error
}

type Repository interface {
	Shorten(context.Context, model.Storage) (string, error)
	Unshorten(context.Context, model.Storage) (string, error)
	GetUrlsByUserID(context.Context, string) ([]model.UrlsStorage, error)
	PingContext(context.Context) error
	helpers
}

type UserService interface {
	helpers
	Shorten(context.Context, model.CreateShortURLInput) (string, error)
	Unshorten(context.Context, model.GetShortURLInput) (string, error)
	GetUrlsByUserID(context.Context, string) ([]model.UrlsResponse, error)
	PingContext(context.Context) error
}
