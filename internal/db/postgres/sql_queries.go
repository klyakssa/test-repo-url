package postgres

import (
	"context"
	"errors"
	"fmt"
	urls "net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/klyakssa/test-repo-url/internal/model"
)

var (
	ErrInsertUniqueViolation = errors.New("insert unique violation")
)

func (s *PostgresStorage) Shorten(ctx context.Context, url model.Storage) (string, error) {
	uuid := uuid.NewString()
	shurl, err := urls.JoinPath(s.cfg.WebConfig.BaseURL, uuid)
	if err != nil {
		return "", fmt.Errorf("join path: %w", err)
	}
	_, err = s.ExecContext(ctx, "INSERT INTO shorten_url (uuid, short_url, original_url, user_id) VALUES ($1, $2, $3, $4)", uuid, shurl, url.OriginalURL, url.UserID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation && strings.Contains(pgErr.ConstraintName, "shorten_url_original_url_key") {
			shurl, err2 := s.selectShortURLByOriginalURL(url.OriginalURL, ctx)
			if err2 != nil {
				return "", err2
			}
			return shurl, fmt.Errorf("%w: URL %q already exists (%w)", ErrInsertUniqueViolation, url, err)
		}
		return "", err
	}
	return shurl, nil
}

func (s *PostgresStorage) Unshorten(ctx context.Context, uuid model.Storage) (url string, err error) {
	err = s.QueryRowContext(ctx, "SELECT original_url FROM shorten_url WHERE uuid = $1", uuid.UUID).Scan(&url)
	if err != nil {
		return "", fmt.Errorf("unshorten: %w", err)
	}
	return url, nil
}

func (s *PostgresStorage) GetUrlsByUserID(ctx context.Context, userID string) ([]model.UrlsStorage, error) {
	var urls []model.UrlsStorage
	err := s.SelectContext(ctx, &urls, "SELECT short_url, original_url FROM shorten_url WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("get urls by user id: %w", err)
	}
	return urls, nil
}

func (s *PostgresStorage) selectShortURLByOriginalURL(url string, ctx context.Context) (shrtURL string, err error) {
	err = s.QueryRowContext(ctx, "SELECT short_url FROM shorten_url WHERE original_url = $1", url).Scan(&shrtURL)
	if err != nil {
		return "", fmt.Errorf("selectShortURLByOriginalURL: %w", err)
	}
	return shrtURL, nil
}
