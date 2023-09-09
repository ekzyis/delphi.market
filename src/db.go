package main

import (
	"database/sql"
	"log"
	"time"

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

type FetchOrdersWhere struct {
	MarketId  int
	Pubkey    string
	Confirmed bool
}

func (db *DB) FetchOrders(where *FetchOrdersWhere, orders *[]Order) error {
	query := "" +
		"SELECT o.id, share_id, o.pubkey, o.side, o.quantity, o.price, o.invoice_id, s.description, s.market_id, i.confirmed_at, o.order_id " +
		"FROM orders o " +
		"JOIN invoices i ON o.invoice_id = i.id " +
		"JOIN shares s ON o.share_id = s.id " +
		"WHERE "
	var args []any
	if where.MarketId > 0 {
		query += "share_id = ANY(SELECT id FROM shares WHERE market_id = $1) "
		args = append(args, where.MarketId)
	} else if where.Pubkey != "" {
		query += "o.pubkey = $1 "
		args = append(args, where.Pubkey)
	}
	if where.Confirmed {
		query += "AND i.confirmed_at IS NOT NULL "
	}
	query += "AND (i.confirmed_at IS NOT NULL OR i.expires_at > CURRENT_TIMESTAMP) "
	query += "ORDER BY price DESC"
	rows, err := db.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var order Order
		rows.Scan(&order.Id, &order.ShareId, &order.Pubkey, &order.Side, &order.Quantity, &order.Price, &order.InvoiceId, &order.Share.Description, &order.Share.MarketId, &order.Invoice.ConfirmedAt, &order.OrderId)
		*orders = append(*orders, order)
	}
	return nil
}

func (db *DB) CreateOrder(order *Order) error {
	if _, err := db.Exec(""+
		"INSERT INTO orders(share_id, pubkey, side, quantity, price, invoice_id) "+
		"VALUES ($1, $2, $3, $4, $5, $6)",
		order.ShareId, order.Pubkey, order.Side, order.Quantity, order.Price, order.InvoiceId); err != nil {
		return err
	}
	return nil
}

func (db *DB) CreateInvoice(invoice *Invoice) error {
	if err := db.QueryRow(""+
		"INSERT INTO invoices(pubkey, msats, preimage, hash, bolt11, created_at, expires_at) "+
		"VALUES($1, $2, $3, $4, $5, $6, $7) "+
		"RETURNING id",
		invoice.Pubkey, invoice.Msats, invoice.Preimage, invoice.PaymentHash, invoice.PaymentRequest, invoice.CreatedAt, invoice.ExpiresAt).Scan(&invoice.Id); err != nil {
		panic(err)
	}
	return nil
}

func (db *DB) FetchInvoice(invoiceId string, invoice *Invoice) error {
	if err := db.QueryRow(""+
		"SELECT id, pubkey, msats, preimage, hash, bolt11, created_at, expires_at, confirmed_at, held_since FROM invoices WHERE id = $1", invoiceId).Scan(&invoice.Id, &invoice.Pubkey, &invoice.Msats, &invoice.Preimage, &invoice.PaymentHash, &invoice.PaymentRequest, &invoice.CreatedAt, &invoice.ExpiresAt, &invoice.ConfirmedAt, &invoice.HeldSince); err != nil {
		return err
	}
	return nil
}

type FetchInvoicesWhere struct {
	Expired bool
}

func (db *DB) FetchInvoices(where *FetchInvoicesWhere, invoices *[]Invoice) error {
	query := "" +
		"SELECT id, msats, msats_received, preimage, hash, bolt11, created_at, expires_at, confirmed_at, held_since " +
		"FROM invoices i "
	if where.Expired {
		query += "WHERE i.expires_at <= CURRENT_TIMESTAMP"
	} else {
		query += "WHERE i.expires_at > CURRENT_TIMESTAMP"
	}
	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var inv Invoice
		rows.Scan(&inv.Id, &inv.Msats, &inv.ReceivedMsats, &inv.Preimage, &inv.PaymentHash, &inv.PaymentRequest, &inv.CreatedAt, &inv.ExpiresAt, &inv.ConfirmedAt, &inv.HeldSince)
		*invoices = append(*invoices, inv)
	}
	return nil
}

func (db *DB) ConfirmInvoice(hash string, confirmedAt time.Time, msatsReceived int) error {
	if _, err := db.Exec("UPDATE invoices SET confirmed_at = $2, msats_received = $3 WHERE hash = $1", hash, confirmedAt, msatsReceived); err != nil {
		return err
	}
	return nil
}

func (db *DB) FetchUser(pubkey string, user *User) error {
	return db.QueryRow("SELECT pubkey FROM users WHERE pubkey = $1", pubkey).Scan(&user.Pubkey)
}
