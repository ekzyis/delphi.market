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

func (db *DB) FetchOrderBook(shareId string, orderBook *[]OrderBookEntry) error {
	rows, err := db.Query(""+
		"SELECT share_id, side, price, SUM(quantity)"+
		"FROM orders WHERE share_id = $1"+
		"GROUP BY (share_id, side, price)"+
		"ORDER BY share_id DESC, side DESC, price DESC", shareId)
	if err != nil {
		return err
	}
	defer rows.Close()
	buyOrders := []Order{}
	sellOrders := []Order{}
	for rows.Next() {
		var order Order
		rows.Scan(&order.ShareId, &order.Side, &order.Price, &order.Quantity)
		if order.Side == "BUY" {
			buyOrders = append(buyOrders, Order{Price: order.Price, Quantity: order.Quantity})
		} else {
			sellOrders = append(sellOrders, Order{Price: order.Price, Quantity: order.Quantity})
		}
	}
	buySum := 0
	sellSum := 0
	for i := 0; i < Max(len(buyOrders), len(sellOrders)); i++ {
		buyPrice, buyQuantity, sellQuantity, sellPrice := 0, 0, 0, 0
		if i < len(buyOrders) {
			buyPrice = buyOrders[i].Price
			buyQuantity = buySum + buyOrders[i].Quantity
		}
		if i < len(sellOrders) {
			sellPrice = sellOrders[i].Price
			sellQuantity = sellSum + sellOrders[i].Quantity
		}
		buySum += buyQuantity
		sellSum += sellQuantity
		*orderBook = append(
			*orderBook,
			OrderBookEntry{
				BuyQuantity:  buyQuantity,
				BuyPrice:     buyPrice,
				SellPrice:    sellPrice,
				SellQuantity: sellQuantity,
			},
		)
	}
	return nil
}
