package db

import (
	"database/sql"
	"log"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/namsral/flag"
)

var (
	db *DB
)

type DB struct {
	*sql.DB
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("error loading env vars: %s", err)
	}
	var dbUrl string
	flag.StringVar(&dbUrl, "DATABASE_URL", "", "Database URL")
	flag.Parse()
	if dbUrl == "" {
		log.Fatal("DATABASE_URL not set")
	}
	db = initDB(dbUrl)
}

func initDB(url string) *DB {
	db, err := sql.Open("postgres", url)
	if err != nil {
		log.Fatal(err)
	}
	// test connection
	_, err = db.Exec("SELECT 1")
	if err != nil {
		log.Fatal(err)
	}
	// TODO: run migrations
	return &DB{DB: db}
}
