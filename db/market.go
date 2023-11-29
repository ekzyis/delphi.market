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
		"INSERT INTO markets(description, end_date, invoice_id) "+
		"VALUES($1, $2, $3) "+
		"RETURNING id", market.Description, market.EndDate, market.InvoiceId).Scan(&market.Id); err != nil {
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
		"SELECT o.id, share_id, o.pubkey, o.side, o.quantity, o.price, o.invoice_id, o.created_at, s.description, s.market_id, i.confirmed_at " +
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
		rows.Scan(&order.Id, &order.ShareId, &order.Pubkey, &order.Side, &order.Quantity, &order.Price, &order.InvoiceId, &order.CreatedAt, &order.Share.Description, &order.Share.MarketId, &order.Invoice.ConfirmedAt)
		*orders = append(*orders, order)
	}
	return nil
}

func (db *DB) CreateOrder(tx *sql.Tx, ctx context.Context, order *Order) error {
	if _, err := tx.ExecContext(ctx, ""+
		"INSERT INTO orders(share_id, pubkey, side, quantity, price, invoice_id) "+
		"VALUES ($1, $2, $3, $4, $5, CASE WHEN $6 = '' THEN NULL ELSE $6::UUID END)",
		order.ShareId, order.Pubkey, order.Side, order.Quantity, order.Price, order.InvoiceId); err != nil {
		return err
	}
	return nil
}

func (db *DB) FetchOrder(tx *sql.Tx, ctx context.Context, orderId string, order *Order) error {
	query := "" +
		"SELECT o.id, o.share_id, o.pubkey, o.side, o.quantity, o.price, o.invoice_id, o.created_at, s.description, s.market_id, i.confirmed_at " +
		"FROM orders o " +
		"JOIN invoices i ON o.invoice_id = i.id " +
		"JOIN shares s ON o.share_id = s.id " +
		"WHERE o.id = $1"
	return tx.QueryRowContext(ctx, query, orderId).Scan(&order.Id, &order.ShareId, &order.Pubkey, &order.Side, &order.Quantity, &order.Price, &order.InvoiceId, &order.CreatedAt, &order.Share.Description, &order.MarketId, &order.Invoice.ConfirmedAt)
}

func (db *DB) FetchUserOrders(pubkey string, orders *[]Order) error {
	query := "" +
		"SELECT o.id, share_id, o.pubkey, o.side, o.quantity, o.price, o.invoice_id, o.created_at, s.description, s.market_id, i.confirmed_at, " +
		"CASE WHEN o.order_id IS NOT NULL THEN 'EXECUTED' ELSE 'PENDING' END AS status " +
		"FROM orders o " +
		"JOIN invoices i ON o.invoice_id = i.id " +
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
		rows.Scan(&order.Id, &order.ShareId, &order.Pubkey, &order.Side, &order.Quantity, &order.Price, &order.InvoiceId, &order.CreatedAt, &order.ShareDescription, &order.Share.MarketId, &order.Invoice.ConfirmedAt, &order.Status)
		*orders = append(*orders, order)
	}
	return nil
}

func (db *DB) FetchMarketOrders(marketId int64, orders *[]Order) error {
	query := "" +
		"SELECT o.id, share_id, o.pubkey, o.side, o.quantity, o.price, o.invoice_id, o.created_at, s.description, s.market_id, " +
		"CASE WHEN o.order_id IS NOT NULL THEN 'EXECUTED' ELSE 'PENDING' END AS status, o.order_id " +
		"FROM orders o " +
		"JOIN shares s ON o.share_id = s.id " +
		"LEFT JOIN invoices i ON i.id = o.invoice_id " +
		"WHERE s.market_id = $1 AND ( (o.side = 'BUY' AND i.confirmed_at IS NOT NULL) OR o.side = 'SELL' ) " +
		"ORDER BY o.created_at DESC"
	rows, err := db.Query(query, marketId)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var order Order
		rows.Scan(&order.Id, &order.ShareId, &order.Pubkey, &order.Side, &order.Quantity, &order.Price, &order.InvoiceId, &order.CreatedAt, &order.ShareDescription, &order.Share.MarketId, &order.Status, &order.OrderId)
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
	if _, err = tx.ExecContext(ctx, "UPDATE orders SET order_id = $1 WHERE id = $2", o1.Id, o2.Id); err != nil {
		log.Println(err)
		tx.Rollback()
		return
	}
	if _, err = tx.ExecContext(ctx, "UPDATE orders SET order_id = $1 WHERE id = $2", o2.Id, o1.Id); err != nil {
		log.Println(err)
		tx.Rollback()
		return
	}
	log.Printf("Matched orders: %s <> %s\n", o1.Id, o2.Id)
	tx.Commit()
}

func (db *DB) FindOrderMatches(tx *sql.Tx, ctx context.Context, o1 *Order, o2 *Order) error {
	query := "" +
		"SELECT o.id FROM orders o " +
		"JOIN shares s ON s.id = o.share_id " +
		"JOIN invoices i ON i.id = o.invoice_id " +
		"WHERE i.confirmed_at IS NOT NULL " +
		"AND o.order_id IS NULL AND o.pubkey <> $1 AND o.quantity = $2 AND s.market_id = $3 AND " +
		"( (o.share_id <> $4 AND o.side = $5::ORDER_SIDE AND o.price = (100 - $6)) OR (o.share_id = $4 AND o.side <> $5::ORDER_SIDE AND o.price = $6)) " +
		"ORDER BY o.created_at ASC LIMIT 1"
	return tx.QueryRowContext(ctx, query, o1.Pubkey, o1.Quantity, o1.Share.MarketId, o1.ShareId, o1.Side, o1.Price).Scan(&o2.Id)
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
		"JOIN invoices i ON i.id = o.invoice_id " +
		"JOIN shares s ON s.id = o.share_id " +
		"WHERE o.pubkey = $1 AND s.market_id = $2 AND ( (o.side = 'BUY' AND i.confirmed_at IS NOT NULL AND o.order_id IS NOT NULL) OR o.side = 'SELL' ) " +
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
