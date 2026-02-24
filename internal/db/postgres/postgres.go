package postgres

import (
	"context"
	"embed"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/klyakssa/test-repo-url/internal/config"
	"github.com/klyakssa/test-repo-url/internal/logger"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

var (
	ErrConnection         = errors.New("connection error")
	ErrMigration          = errors.New("migration error")
	ErrNoConnectionString = errors.New("no connection string")
)

type PostgresStorage struct {
	*sqlx.DB
	cfg *config.Config
	l   *logger.MyLogger
}

func NewPostgresStorage(cfg *config.Config, l *logger.MyLogger) (*PostgresStorage, error) {
	c, err := connectPostgres(cfg)
	return &PostgresStorage{
			c,
			cfg,
			l,
		},
		err
}

func (s *PostgresStorage) PingContext(ctx context.Context) error {
	return s.DB.PingContext(ctx)
}

func (s *PostgresStorage) Close() error {
	return s.DB.Close()
}

func connectPostgres(cfg *config.Config) (*sqlx.DB, error) {
	if cfg.PostDB.ConnString == "" {
		return nil, ErrNoConnectionString
	}
	dbpool, err := pgxpool.New(context.Background(), cfg.PostDB.ConnString)
	if err != nil {
		return nil, fmt.Errorf("pgxpool: %w", err)
	}
	db := stdlib.OpenDBFromPool(dbpool)
	sqlxDB := sqlx.NewDb(db, "pgx")
	if err := applyMigrations(sqlxDB); err != nil {
		return nil, fmt.Errorf("apply migrations: %w", err)
	}
	return sqlxDB, nil
}

func applyMigrations(db *sqlx.DB) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	if err := goose.Up(db.DB, "migrations"); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	return nil
}
