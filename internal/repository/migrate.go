package repository

import (
	"embed"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migration/*.sql
var embedMigrations embed.FS

func migrate(pool *pgxpool.Pool) error {
	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("goose.SetDialect: %w", err)
	}

	db := stdlib.OpenDBFromPool(pool)
	if err := goose.Up(db, "migration"); err != nil {
		return fmt.Errorf("goose.Up: %w", err)
	}

	if err := db.Close(); err != nil {
		return fmt.Errorf("migrate.db.Close: %w", err)
	}

	return nil
}
