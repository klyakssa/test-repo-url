package repository

import "context"

type helpers interface {
	Close() error
}

type Repository interface {
	UserService
	helpers
}

type UserService interface {
	helpers
	Shorten(context.Context, string) (string, error)
	Unshorten(context.Context, string) (string, error)
	PingContext(context.Context) error
}
