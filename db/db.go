package db

import (
	"database/sql"
	"fmt"
	"io/ioutil"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

var (
	initSqlPath = "./db/init.sql"
)

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

func (db *DB) Reset(dbName string) error {
	var (
		f   []byte
		err error
	)
	if err = db.Clear(dbName); err != nil {
		return err
	}
	if f, err = ioutil.ReadFile(initSqlPath); err != nil {
		return err
	}
	if _, err = db.Exec(string(f)); err != nil {
		return err
	}
	return nil
}

func (db *DB) Clear(dbName string) error {
	var (
		tables = []string{"lnauth", "users", "sessions", "markets", "shares", "invoices", "order_side", "orders", "matches"}
		sql    []string
		err    error
	)
	for _, t := range tables {
		sql = append(sql, fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", t))
	}
	sql = append(sql, "DROP EXTENSION IF EXISTS \"uuid-ossp\"")
	sql = append(sql, "DROP TYPE IF EXISTS order_side")
	for _, s := range sql {
		if _, err = db.Exec(s); err != nil {
			return err
		}
	}
	return nil
}
