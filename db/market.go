package db

import (
	"context"
	"database/sql"
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
		"VALUES ($1, $2, $3, $4, $5, $6)",
		order.ShareId, order.Pubkey, order.Side, order.Quantity, order.Price, order.InvoiceId); err != nil {
		return err
	}
	return nil
}

func (db *DB) FetchUserOrders(pubkey string, orders *[]Order) error {
	query := "" +
		"SELECT o.id, share_id, o.pubkey, o.side, o.quantity, o.price, o.invoice_id, o.created_at, s.description, s.market_id, i.confirmed_at " +
		"FROM orders o " +
		"JOIN invoices i ON o.invoice_id = i.id " +
		"JOIN shares s ON o.share_id = s.id " +
		"WHERE o.pubkey = $1 AND i.confirmed_at IS NOT NULL " +
		"ORDER BY o.created_at DESC"
	rows, err := db.Query(query, pubkey)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var order Order
		rows.Scan(&order.Id, &order.ShareId, &order.Pubkey, &order.Side, &order.Quantity, &order.Price, &order.InvoiceId, &order.CreatedAt, &order.ShareDescription, &order.Share.MarketId, &order.Invoice.ConfirmedAt)
		*orders = append(*orders, order)
	}
	return nil
}

func (db *DB) FetchMarketOrders(marketId int64, orders *[]Order) error {
	query := "" +
		"SELECT o.id, share_id, o.pubkey, o.side, o.quantity, o.price, o.invoice_id, o.created_at, s.description, s.market_id " +
		"FROM orders o " +
		"JOIN shares s ON o.share_id = s.id " +
		"JOIN invoices i ON i.id = o.invoice_id " +
		"WHERE s.market_id = $1 AND i.confirmed_at IS NOT NULL " +
		"ORDER BY o.created_at DESC"
	rows, err := db.Query(query, marketId)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var order Order
		rows.Scan(&order.Id, &order.ShareId, &order.Pubkey, &order.Side, &order.Quantity, &order.Price, &order.InvoiceId, &order.CreatedAt, &order.ShareDescription, &order.Share.MarketId)
		*orders = append(*orders, order)
	}
	return nil
}
