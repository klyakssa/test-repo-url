package postgres

import (
	"context"
	"errors"
	"fmt"
	urls "net/url"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrInsertUniqueViolation = errors.New("insert unique violation")
)

func (s *PostgresStorage) Shorten(ctx context.Context, url string) (string, error) {
	uuid := uuid.NewString()
	shurl, err := urls.JoinPath(s.cfg.WebConfig.BaseURL, uuid)
	if err != nil {
		return "", fmt.Errorf("join path: %w", err)
	}
	_, err = s.ExecContext(ctx, "INSERT INTO shorten_url (uuid, short_url, original_url) VALUES ($1, $2, $3)", uuid, shurl, url)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			shurl, err2 := s.selectShortURLByOriginalURL(url, ctx)
			if err2 != nil {
				return "", err2
			}
			return shurl, fmt.Errorf("%w: URL %q already exists (%w)", ErrInsertUniqueViolation, url, err)
		}
		return "", err
	}
	return shurl, nil
}

func (s *PostgresStorage) Unshorten(ctx context.Context, uuid string) (url string, err error) {
	err = s.QueryRowContext(ctx, "SELECT original_url FROM shorten_url WHERE uuid = $1", uuid).Scan(&url)
	if err != nil {
		return "", fmt.Errorf("unshorten: %w", err)
	}
	return url, nil
}

func (s *PostgresStorage) selectShortURLByOriginalURL(url string, ctx context.Context) (shrtURL string, err error) {
	err = s.QueryRowContext(ctx, "SELECT short_url FROM shorten_url WHERE original_url = $1", url).Scan(&shrtURL)
	if err != nil {
		return "", fmt.Errorf("selectShortURLByOriginalURL: %w", err)
	}
	return shrtURL, nil
}
