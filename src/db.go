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
	db    *DB
)

type DB struct {
	*sql.DB
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	flag.StringVar(&DbUrl, "DATABASE_URL", "", "Database URL")
	flag.Parse()
	validateFlags()
	db = initDb()
	_, err = db.Exec("SELECT 1")
	if err != nil {
		log.Fatal(err)
	}
}

func initDb() *DB {
	db, err := sql.Open("postgres", DbUrl)
	if err != nil {
		log.Fatal(err)
	}
	return &DB{DB: db}
}

func validateFlags() {
	if DbUrl == "" {
		log.Fatal("DATABASE_URL not set")
	}
}

func (db *DB) FetchMarket(marketId int, market *Market) error {
	if err := db.QueryRow("SELECT id, description FROM markets WHERE id = $1", marketId).Scan(&market.Id, &market.Description); err != nil {
		return err
	}
	return nil
}

func (db *DB) FetchShares(marketId int, shares *[]Share) error {
	rows, err := db.Query("SELECT id, market_id, description FROM shares WHERE market_id = $1 ORDER BY description DESC", marketId)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var share Share
		rows.Scan(&share.Id, &share.MarketId, &share.Description)
		*shares = append(*shares, share)
	}
	return nil
}

func (db *DB) FetchOrders(marketId int, orders *[]Order) error {
	rows, err := db.Query(""+
		"SELECT id, share_id, pubkey, side, quantity, price, order_id FROM orders "+
		"WHERE share_id = ANY(SELECT id FROM shares WHERE market_id = $1) "+
		"ORDER BY price DESC", marketId)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var order Order
		rows.Scan(&order.Id, &order.ShareId, &order.Pubkey, &order.Side, &order.Quantity, &order.Price, &order.OrderId)
		*orders = append(*orders, order)
	}
	return nil
}
