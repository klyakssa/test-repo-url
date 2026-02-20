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
	Shorten(string, context.Context) (string, error)
	Unshorten(string, context.Context) (string, error)
	PingContext(context.Context) error
}
