package database

import (
	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/henry40408/lmb"
	_ "github.com/mattn/go-sqlite3"
)

func migrateDB(db *sql.DB) error {
	d, err := iofs.New(lmb.MigrationFiles, "migrations")
	if err != nil {
		return err
	}
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithInstance("iofs", d, "sqlite", driver)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

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
	err = migrateDB(db)
	if err != nil {
		return nil, err
	}
	return db, nil
}
