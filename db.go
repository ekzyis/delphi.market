package main

import (
	"database/sql"
	"log"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/namsral/flag"
)

var (
	DbUrl string
	db    *sql.DB
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	flag.StringVar(&DbUrl, "DATABASE_URL", "", "Database URL")
	flag.Parse()
	validateFlags()
	db = initDb()
}

func initDb() *sql.DB {
	db, err := sql.Open("postgres", DbUrl)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func validateFlags() {
	if DbUrl == "" {
		log.Fatal("DATABASE_URL not set")
	}
}
