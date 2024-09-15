package store

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"reflect"
	"unsafe"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/henry40408/lmb"
	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	db *sql.DB
}

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

func NewStore(dsn string) (Store, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return Store{}, err
	}
	db.SetMaxOpenConns(1)
	_, err = db.Exec(`
    PRAGMA busy_timeout = 5000;
    PRAGMA foreign_keys = OFF;
    PRAGMA journal_mode = wal;
    PRAGMA synchronous = NORMAL;
  `)
	if err != nil {
		return Store{}, err
	}
	err = migrateDB(db)
	if err != nil {
		return Store{}, err
	}
	return Store{db}, nil
}

func deserializeData(value []byte, target interface{}) error {
	decoder := gob.NewDecoder(bytes.NewBuffer(value))
	return decoder.Decode(target)
}

func (s *Store) Get(name string) (interface{}, error) {
	stmt, err := s.db.Prepare(`SELECT value FROM store WHERE name = ?`)
	if err != nil {
		return nil, err
	}
	var value []byte
	err = stmt.QueryRow(&name).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			return nil, err
		}
	}
	var deserialized interface{}
	err = deserializeData(value, &deserialized)
	if err != nil {
		return nil, err
	}
	return deserialized, nil
}

func serializeData(data interface{}) []byte {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	encoder.Encode(data)
	return buffer.Bytes()
}

func (s *Store) Put(name string, value interface{}) error {
	stmt, err := s.db.Prepare(`INSERT OR REPLACE INTO store (name, value, type_hint, size) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	serialized := serializeData(&value)
	_, err = stmt.Exec(&name, serialized, reflect.TypeOf(value).Name(), int64(unsafe.Sizeof(value)))
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) Begin() (*StoreTx, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	return &StoreTx{tx}, nil
}

type StoreTx struct {
	tx *sql.Tx
}

func (st *StoreTx) Rollback() error {
	return st.tx.Rollback()
}

func (st *StoreTx) Commit() error {
	return st.tx.Commit()
}

func (st *StoreTx) Get(name string) (interface{}, error) {
	stmt, err := st.tx.Prepare(`SELECT value FROM store WHERE name = ?`)
	if err != nil {
		return nil, err
	}
	var value []byte
	err = stmt.QueryRow(&name).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			return nil, err
		}
	}
	var deserialized interface{}
	err = deserializeData(value, &deserialized)
	if err != nil {
		return nil, err
	}
	return deserialized, nil

}

func (st *StoreTx) Put(name string, value interface{}) error {
	stmt, err := st.tx.Prepare(`INSERT OR REPLACE INTO store (name, value, type_hint, size) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	serialized := serializeData(&value)
	_, err = stmt.Exec(&name, serialized, reflect.TypeOf(value).Name(), int64(unsafe.Sizeof(value)))
	if err != nil {
		return err
	}
	return nil
}
