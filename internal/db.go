package internal

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func OpenDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	_, err = db.Exec(`
  PRAGMA busy_timeout = 5000;
  PRAGMA foreign_keys = OFF;
  PRAGMA journal_mode = wal;
  PRAGMA synchronous = NORMAL;
  `)
	if err != nil {
		return nil, err
	}
	return db, nil
}
