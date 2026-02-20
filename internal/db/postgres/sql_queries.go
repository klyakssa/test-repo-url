package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (s *PostgresStorage) Shorten(url string, ctx context.Context) (string, error) {
	uuid := uuid.NewString()
	shurl := fmt.Sprintf("%s/%s", s.cfg.WebConfig.BaseUrl, uuid)
	_, err := s.ExecContext(ctx, "INSERT INTO shorten_url (uuid, short_url, original_url) VALUES ($1, $2, $3)", uuid, shurl, url)
	if err != nil {
		return "", err
	}
	return shurl, nil
}

func (s *PostgresStorage) Unshorten(uuid string, ctx context.Context) (url string, err error) {
	err = s.QueryRowContext(ctx, "SELECT original_url FROM shorten_url WHERE uuid = $1", uuid).Scan(&url)
	return url, err
}
