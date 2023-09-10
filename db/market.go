package db

import "database/sql"

func FetchMarket(marketId int, market *Market) error {
	if err := db.QueryRow("SELECT id, description FROM markets WHERE id = $1", marketId).Scan(&market.Id, &market.Description); err != nil {
		return err
	}
	return nil
}

func FetchActiveMarkets(markets *[]Market) error {
	var (
		rows   *sql.Rows
		market Market
		err    error
	)
	if rows, err = db.Query("SELECT id, description, active FROM markets WHERE active = true"); err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&market.Id, &market.Description, &market.Active)
		*markets = append(*markets, market)
	}
	return nil
}

func FetchShares(marketId int, shares *[]Share) error {
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

func FetchOrders(where *FetchOrdersWhere, orders *[]Order) error {
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

func CreateOrder(order *Order) error {
	if _, err := db.Exec(""+
		"INSERT INTO orders(share_id, pubkey, side, quantity, price, invoice_id) "+
		"VALUES ($1, $2, $3, $4, $5, $6)",
		order.ShareId, order.Pubkey, order.Side, order.Quantity, order.Price, order.InvoiceId); err != nil {
		return err
	}
	return nil
}
