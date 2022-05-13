package db

import (
	"database/sql"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"go.uber.org/zap"
	"time"
)

// Open opens a postgres database for the provided connection details.
func Open(conn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", conn)
	if err != nil {
		return nil, err
	}

	if err = checkHealth(db); err != nil {
		return nil, err
	}
	zap.L().Debug("database connection opened")
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
