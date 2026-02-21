package postgres

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"os"

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
	ErrConnection = errors.New("connection error")
	ErrMigration  = errors.New("migration error")
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

func (ps *PostgresStorage) PingContext(ctx context.Context) error {
	return ps.PingContext(ctx)
}

func (ps *PostgresStorage) Close() error {
	return ps.DB.Close()
}

func connectPostgres(cfg *config.Config) (*sqlx.DB, error) {
	var errs []error
	dbpool, err := pgxpool.New(context.Background(), cfg.PostDB.ConnString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		errs = append(errs, err)
	}
	db := stdlib.OpenDBFromPool(dbpool)
	sqlxDB := sqlx.NewDb(db, "pgx")
	if err := applyMigrations(sqlxDB); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to apply migration: %v\n", err)
		errs = append(errs, err)
	}
	return sqlxDB, errors.Join(errs...)
}

func applyMigrations(db *sqlx.DB) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.Up(db.DB, "migrations"); err != nil {
		return err
	}

	return nil
}
