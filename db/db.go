package db

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

func New(dbUrl string) (*DB, error) {
	var (
		db_ *sql.DB
		db  *DB
		err error
	)
	if db_, err = sql.Open("postgres", dbUrl); err != nil {
		return nil, err
	}
	// test connection
	if _, err = db_.Exec("SELECT 1"); err != nil {
		return nil, err
	}
	// TODO: run migrations
	db = &DB{DB: db_}
	return db, nil
}
