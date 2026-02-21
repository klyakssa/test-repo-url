package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrInsertUniqueViolation = errors.New("insert unique violation")
)

func (s *PostgresStorage) Shorten(url string, ctx context.Context) (string, error) {
	uuid := uuid.NewString()
	shurl := fmt.Sprintf("%s/%s", s.cfg.WebConfig.BaseURL, uuid)
	_, err := s.ExecContext(ctx, "INSERT INTO shorten_url (uuid, short_url, original_url) VALUES ($1, $2, $3)", uuid, shurl, url)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			shurl, err = s.selectShortURLByOriginalURL(url, ctx)
			if err != nil {
				s.l.Error("failed to fetch existing record after unique violation", "error", err)
				return "", err
			}
			return shurl, ErrInsertUniqueViolation
		}
		return "", err
	}
	return shurl, nil
}

func (s *PostgresStorage) Unshorten(uuid string, ctx context.Context) (url string, err error) {
	err = s.QueryRowContext(ctx, "SELECT original_url FROM shorten_url WHERE uuid = $1", uuid).Scan(&url)
	return url, err
}

func (s *PostgresStorage) selectShortURLByOriginalURL(url string, ctx context.Context) (shrtURL string, err error) {
	err = s.QueryRowContext(ctx, "SELECT short_url FROM shorten_url WHERE original_url = $1", url).Scan(&shrtURL)
	return shrtURL, err
}
