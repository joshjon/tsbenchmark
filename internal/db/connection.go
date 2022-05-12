package db

import (
	"database/sql"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"time"
)

func Open(conn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", conn)
	if err != nil {
		return nil, err
	}

	if err = checkHealth(db); err != nil {
		return nil, err
	}
	return db, nil
}

func checkHealth(db *sql.DB) error {
	retries := 10
	interval := time.Second

	var err error

	for i := 0; i < retries; i++ {
		if err = db.Ping(); err == nil {
			return nil
		}
		time.Sleep(interval)
	}

	return err
}
