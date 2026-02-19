package postgres

import (
	"context"
	"embed"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/klyakssa/test-repo-url/internal/config"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

type PostgresDB struct {
}

func ConnectPostgres(cfg *config.Config) (*sqlx.DB, error) {
	dbpool, err := pgxpool.New(context.Background(), cfg.PostDB.ConnString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	db := stdlib.OpenDBFromPool(dbpool)
	sqlxDB := sqlx.NewDb(db, "pgx")
	if err := applyMigrations(sqlxDB); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to apply migration: %v\n", err)
		return nil, err
	}
	return sqlxDB, nil
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
