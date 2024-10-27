package repository

import "github.com/jackc/pgx/v5/pgxpool"

type DatabaseRepository struct {
	db *pgxpool.Pool
}

func (dr *DatabaseRepository) Close() error {
	dr.db.Close()
	return nil
}
