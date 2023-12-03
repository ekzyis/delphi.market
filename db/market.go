package db

import (
	"context"
	"database/sql"
	"log"
	"time"
)

type FetchOrdersWhere struct {
	MarketId  int
	Pubkey    string
	Confirmed bool
}

func (db *DB) CreateMarket(tx *sql.Tx, ctx context.Context, market *Market) error {
	if err := tx.QueryRowContext(ctx, ""+
		"INSERT INTO markets(description, end_date, pubkey, invoice_id) "+
		"VALUES($1, $2, $3, $4) "+
		"RETURNING id", market.Description, market.EndDate, market.Pubkey, market.InvoiceId).Scan(&market.Id); err != nil {
		return err
	}
	// For now, we only support binary markets.
	if _, err := tx.Exec("INSERT INTO shares(market_id, description) VALUES ($1, 'YES'), ($1, 'NO')", market.Id); err != nil {
		return err
	}
	return nil
}

func (db *DB) FetchMarket(marketId int, market *Market) error {
	if err := db.QueryRow("SELECT id, description, end_date FROM markets WHERE id = $1", marketId).Scan(&market.Id, &market.Description, &market.EndDate); err != nil {
		return err
	}
	return nil
}

func (db *DB) FetchActiveMarkets(markets *[]Market) error {
	var (
		rows   *sql.Rows
		market Market
		err    error
	)
	if rows, err = db.Query("" +
		"SELECT m.id, m.description, m.end_date FROM markets m " +
		"JOIN invoices i ON i.id = m.invoice_id WHERE i.confirmed_at IS NOT NULL"); err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&market.Id, &market.Description, &market.EndDate)
		*markets = append(*markets, market)
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

func (db *DB) FetchShare(tx *sql.Tx, ctx context.Context, shareId string, share *Share) error {
	return tx.QueryRowContext(ctx, "SELECT id, market_id, description FROM shares WHERE id = $1", shareId).Scan(&share.Id, &share.MarketId, &share.Description)
}

func (db *DB) FetchOrders(where *FetchOrdersWhere, orders *[]Order) error {
	query := "" +
		"SELECT o.id, share_id, o.pubkey, o.side, o.quantity, o.price, o.invoice_id, o.created_at, o.deleted_at, s.description, s.market_id, i.confirmed_at " +
		"FROM orders o " +
		"JOIN invoices i ON o.invoice_id = i.id " +
		"JOIN shares s ON o.share_id = s.id " +
		"WHERE o.deleted_at IS NULL "
	var args []any
	if where.MarketId > 0 {
		query += "AND share_id = ANY(SELECT id FROM shares WHERE market_id = $1) "
		args = append(args, where.MarketId)
	} else if where.Pubkey != "" {
		query += "AND o.pubkey = $1 "
		args = append(args, where.Pubkey)
	}
	if where.Confirmed {
		query += "AND o.order_id IS NOT NULL "
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
		rows.Scan(&order.Id, &order.ShareId, &order.Pubkey, &order.Side, &order.Quantity, &order.Price, &order.InvoiceId, &order.CreatedAt, &order.DeletedAt, &order.Share.Description, &order.Share.MarketId, &order.Invoice.ConfirmedAt)
		*orders = append(*orders, order)
	}
	return nil
}

func (db *DB) CreateOrder(tx *sql.Tx, ctx context.Context, order *Order) error {
	if _, err := tx.ExecContext(ctx, ""+
		"INSERT INTO orders(share_id, pubkey, side, quantity, price, invoice_id) "+
		"VALUES ($1, $2, $3, $4, $5, CASE WHEN $6 = '' THEN NULL ELSE $6::UUID END)",
		order.ShareId, order.Pubkey, order.Side, order.Quantity, order.Price, order.InvoiceId.String); err != nil {
		return err
	}
	return nil
}

func (db *DB) FetchOrder(tx *sql.Tx, ctx context.Context, orderId string, order *Order) error {
	query := "" +
		"SELECT o.id, o.share_id, o.pubkey, o.side, o.quantity, o.price, o.created_at, o.deleted_at, o.order_id, s.description, s.market_id, i.confirmed_at, o.invoice_id, COALESCE(i.msats_received, 0) " +
		"FROM orders o " +
		"LEFT JOIN invoices i ON o.invoice_id = i.id " +
		"JOIN shares s ON o.share_id = s.id " +
		"WHERE o.id = $1"
	return tx.QueryRowContext(ctx, query, orderId).Scan(&order.Id, &order.ShareId, &order.Pubkey, &order.Side, &order.Quantity, &order.Price, &order.CreatedAt, &order.DeletedAt, &order.OrderId, &order.Share.Description, &order.MarketId, &order.Invoice.ConfirmedAt, &order.InvoiceId, &order.Invoice.MsatsReceived)
}

func (db *DB) FetchUserOrders(pubkey string, orders *[]Order) error {
	query := "" +
		"SELECT o.id, share_id, o.pubkey, o.side, o.quantity, o.price, o.created_at, o.deleted_at, s.description, s.market_id, i.confirmed_at, " +
		"CASE WHEN o.order_id IS NOT NULL THEN 'EXECUTED' WHEN o.deleted_at IS NOT NULL THEN 'CANCELED' ELSE 'PENDING' END AS status, o.order_id, o.invoice_id " +
		"FROM orders o " +
		"LEFT JOIN invoices i ON o.invoice_id = i.id " +
		"JOIN shares s ON o.share_id = s.id " +
		"WHERE o.pubkey = $1 AND ( (o.side = 'BUY' AND i.confirmed_at IS NOT NULL) OR o.side = 'SELL' ) " +
		"ORDER BY o.created_at DESC"
	rows, err := db.Query(query, pubkey)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var order Order
		rows.Scan(&order.Id, &order.ShareId, &order.Pubkey, &order.Side, &order.Quantity, &order.Price, &order.CreatedAt, &order.DeletedAt, &order.ShareDescription, &order.Share.MarketId, &order.Invoice.ConfirmedAt, &order.Status, &order.OrderId, &order.InvoiceId)
		*orders = append(*orders, order)
	}
	return nil
}

func (db *DB) FetchMarketOrders(marketId int64, orders *[]Order) error {
	query := "" +
		"SELECT o.id, share_id, o.pubkey, o.side, o.quantity, o.price, o.created_at, s.description, s.market_id, " +
		"CASE WHEN o.order_id IS NOT NULL THEN 'EXECUTED' ELSE 'PENDING' END AS status, o.order_id, o.invoice_id " +
		"FROM orders o " +
		"JOIN shares s ON o.share_id = s.id " +
		"LEFT JOIN invoices i ON o.invoice_id = i.id " +
		"WHERE s.market_id = $1 AND o.deleted_at IS NULL AND ( (o.side = 'BUY' AND i.confirmed_at IS NOT NULL) OR o.side = 'SELL' ) " +
		"ORDER BY o.created_at DESC"
	rows, err := db.Query(query, marketId)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var order Order
		rows.Scan(&order.Id, &order.ShareId, &order.Pubkey, &order.Side, &order.Quantity, &order.Price, &order.CreatedAt, &order.ShareDescription, &order.Share.MarketId, &order.Status, &order.OrderId, &order.InvoiceId)
		*orders = append(*orders, order)
	}
	return nil
}

func (db *DB) RunMatchmaking(orderId string) {
	var (
		ctx    context.Context
		cancel context.CancelFunc
		tx     *sql.Tx
		o1     Order
		o2     Order
		err    error
	)
	ctx, cancel = context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()
	if tx, err = db.BeginTx(ctx, nil); err != nil {
		log.Println(err)
		return
	}
	// TODO: assert that order was confirmed
	if err = db.FetchOrder(tx, ctx, orderId, &o1); err != nil {
		log.Println(err)
		tx.Rollback()
		return
	}
	if err = db.FindOrderMatches(tx, ctx, &o1, &o2); err != nil {
		log.Println(err)
		tx.Rollback()
		return
	}
	if o2.OrderId.Valid {
		log.Printf("assertion failed: order %s matched order %s but order_id already set to %s\n", o1.Id, o2.Id, o2.OrderId.String)
		tx.Rollback()
		return
	}
	if o1.Id == o2.Id {
		log.Printf("assertion failed: order %s matched itself", o1.Id)
		tx.Rollback()
		return
	}
	if err = db.MatchOrders(tx, ctx, &o1, &o2); err != nil {
		log.Println(err)
		tx.Rollback()
		return
	}

	tx.Commit()
}

func (db *DB) FindOrderMatches(tx *sql.Tx, ctx context.Context, o1 *Order, o2 *Order) error {
	query := "" +
		"SELECT o.id, o.order_id, o.side, o.quantity, o.price, o.pubkey FROM orders o " +
		"JOIN shares s ON s.id = o.share_id " +
		"LEFT JOIN invoices i ON i.id = o.invoice_id " +
		// only match orders which are not soft deleted
		"WHERE o.deleted_at IS NULL " +
		// only match orders which are not already settled
		"AND o.order_id IS NULL " +
		// only match orders from other users
		"AND o.pubkey <> $1 " +
		// orders must always be for same market and have same quantity
		"AND o.quantity = $2 AND s.market_id = $3 " +
		// BUY orders must have been confirmed by paying the invoice
		"AND CASE WHEN o.side = 'BUY' THEN i.confirmed_at IS NOT NULL ELSE 1=1 END " +
		"AND (" +
		// -- BUY orders match if they are for different shares and the sum of their prices equal 100
		// -- example: BUY 5 YES @ 60 <> BUY 5 NO @ 40
		"  ( $5 = 'BUY' AND o.side = 'BUY' AND o.price = (100-$6) AND o.share_id <> $4 ) " +
		// -- BUY orders match SELL orders if they are for the same share and have same price
		// -- example: BUY 5 YES @ 60 <> SELL 5 YES @ 60
		"  OR ( $5 = 'BUY' AND o.side = 'SELL' AND o.price = $6 AND o.share_id = $4 ) " +
		// -- SELL orders match BUY orders if they are for the same share and have same price
		// -- example: SELL 5 YES @ 60 <> BUY 5 YES @ 60
		"  OR ( $5 = 'SELL' AND o.side = 'BUY' AND o.price = $6 AND o.share_id = $4 ) " +
		") " +
		// match oldest order first
		"ORDER BY o.created_at ASC LIMIT 1"
	return tx.QueryRowContext(ctx, query, o1.Pubkey, o1.Quantity, o1.Share.MarketId, o1.ShareId, o1.Side, o1.Price).Scan(&o2.Id, &o2.OrderId, &o2.Side, &o2.Quantity, &o2.Price, &o2.Pubkey)
}

func (db *DB) MatchOrders(tx *sql.Tx, ctx context.Context, o1 *Order, o2 *Order) error {
	var err error
	if _, err = tx.ExecContext(ctx, "UPDATE orders SET order_id = $1 WHERE id = $2", o1.Id, o2.Id); err != nil {
		tx.Rollback()
		return err
	}
	if _, err = tx.ExecContext(ctx, "UPDATE orders SET order_id = $1 WHERE id = $2", o2.Id, o1.Id); err != nil {
		tx.Rollback()
		return err
	}
	if o1.Side == "SELL" {
		if _, err = tx.ExecContext(ctx, "UPDATE users SET msats = msats + $1 WHERE pubkey = $2", (o1.Price*o1.Quantity)*1000, o1.Pubkey); err != nil {
			tx.Rollback()
			return err
		}
	}
	if o2.Side == "SELL" {
		if _, err = tx.ExecContext(ctx, "UPDATE users SET msats = msats + $1 WHERE pubkey = $2", (o2.Price*o2.Quantity)*1000, o2.Pubkey); err != nil {
			tx.Rollback()
			return err
		}
	}
	log.Printf("Matched orders: %s <> %s\n", o1.Id, o2.Id)
	return nil
}

// [
//
//	{ "x": <timestamp>, "y": { <share_description>: <score>, ... } },
//
// ]
type MarketStat struct {
	X time.Time      `json:"x"`
	Y map[string]int `json:"y"`
}
type MarketStats = []MarketStat

func (db *DB) FetchMarketStats(marketId int64, stats *MarketStats) error {
	query := "" +
		"SELECT " +
		"s.description, " +
		"GREATEST(i.confirmed_at, i2.confirmed_at) AS confirmed_at, " +
		"SUM(o.price * o.quantity) OVER (PARTITION BY o.share_id ORDER BY o.created_at ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW) AS score " +
		"FROM orders o " +
		"JOIN orders o2 ON o2.id = o.order_id " +
		"JOIN shares s ON s.id = o.share_id " +
		"JOIN invoices i ON i.id = o.invoice_id " +
		"JOIN invoices i2 ON i2.id = o2.invoice_id " +
		"WHERE s.market_id = $1 AND i.confirmed_at IS NOT NULL AND o.order_id IS NOT NULL ORDER BY i.confirmed_at ASC"
	rows, err := db.Query(query, marketId)
	if err != nil {
		return err
	}

	defer rows.Close()
	for rows.Next() {
		var stat MarketStat
		var (
			timestamp   time.Time
			description string
			score       int
		)
		rows.Scan(&description, &timestamp, &score)
		stat.X = timestamp
		stat.Y = map[string]int{
			description: score,
		}
		*stats = append(*stats, stat)
	}
	return nil
}

func (db *DB) FetchUserBalance(tx *sql.Tx, ctx context.Context, marketId int, pubkey string, balance *map[string]any) error {
	query := "" +
		"SELECT s.description, " +
		"SUM(CASE WHEN o.side = 'BUY' THEN o.quantity ELSE -o.quantity END) " +
		"FROM orders o " +
		"LEFT JOIN invoices i ON i.id = o.invoice_id " +
		"JOIN shares s ON s.id = o.share_id " +
		"WHERE o.pubkey = $1 AND s.market_id = $2 AND o.deleted_at IS NULL " +
		"AND ( (o.side = 'BUY' AND i.confirmed_at IS NOT NULL AND o.order_id IS NOT NULL) OR o.side = 'SELL' ) " +
		"GROUP BY o.pubkey, s.description"
	rows, err := tx.QueryContext(ctx, query, pubkey, marketId)
	if err != nil {
		return err
	}

	defer rows.Close()
	for rows.Next() {
		var (
			sdesc string
			val   int
		)
		if err = rows.Scan(&sdesc, &val); err != nil {
			return err
		}
		(*balance)[sdesc] = val
	}
	return nil
}
