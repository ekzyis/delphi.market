package db

import (
	"context"
	"database/sql"
	"time"
)

func (db *DB) CreateInvoice(tx *sql.Tx, ctx context.Context, invoice *Invoice) error {
	if err := tx.QueryRowContext(ctx, ""+
		"INSERT INTO invoices(pubkey, msats, preimage, hash, bolt11, created_at, expires_at, description) "+
		"VALUES($1, $2, $3, $4, $5, $6, $7, $8) "+
		"RETURNING id",
		invoice.Pubkey, invoice.Msats, invoice.Preimage, invoice.Hash, invoice.PaymentRequest, invoice.CreatedAt, invoice.ExpiresAt, invoice.Description).Scan(&invoice.Id); err != nil {
		return err
	}
	return nil
}

type FetchInvoiceWhere struct {
	Id   string
	Hash string
}

func (db *DB) FetchInvoice(where *FetchInvoiceWhere, invoice *Invoice) error {
	var (
		query = "SELECT id, pubkey, msats, preimage, hash, bolt11, created_at, expires_at, confirmed_at, held_since, COALESCE(description, '') FROM invoices "
		args  []any
	)
	if where.Id != "" {
		query += "WHERE id = $1"
		args = append(args, where.Id)
	} else if where.Hash != "" {
		query += "WHERE hash = $1"
		args = append(args, where.Hash)
	}
	if err := db.QueryRow(query, args...).Scan(
		&invoice.Id, &invoice.Pubkey, &invoice.Msats, &invoice.Preimage, &invoice.Hash,
		&invoice.PaymentRequest, &invoice.CreatedAt, &invoice.ExpiresAt, &invoice.ConfirmedAt, &invoice.HeldSince, &invoice.Description); err != nil {
		return err
	}
	return nil
}

type FetchInvoicesWhere struct {
	Unconfirmed bool
}

func (db *DB) FetchInvoices(where *FetchInvoicesWhere, invoices *[]Invoice) error {
	var (
		rows    *sql.Rows
		invoice Invoice
		err     error
	)
	var (
		query = "SELECT id, pubkey, msats, preimage, hash, bolt11, created_at, expires_at, confirmed_at, held_since, COALESCE(description, '') FROM invoices "
	)
	if where.Unconfirmed {
		query += "WHERE confirmed_at IS NULL"
	}
	if rows, err = db.Query(query); err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		rows.Scan(
			&invoice.Id, &invoice.Pubkey, &invoice.Msats, &invoice.Preimage, &invoice.Hash,
			&invoice.PaymentRequest, &invoice.CreatedAt, &invoice.ExpiresAt, &invoice.ConfirmedAt, &invoice.HeldSince, &invoice.Description)
		*invoices = append(*invoices, invoice)
	}
	return nil
}

func (db *DB) FetchUserInvoices(pubkey string, invoices *[]Invoice) error {
	var (
		rows    *sql.Rows
		invoice Invoice
		err     error
	)
	var (
		query = "" +
			"SELECT id, pubkey, msats, preimage, hash, bolt11, created_at, expires_at, confirmed_at, held_since, COALESCE(description, ''), " +
			"CASE WHEN confirmed_at IS NOT NULL THEN 'PAID' WHEN expires_at < CURRENT_TIMESTAMP THEN 'EXPIRED' ELSE 'WAITING' END AS status " +
			"FROM invoices " +
			"WHERE pubkey = $1 " +
			"ORDER BY created_at DESC"
	)
	if rows, err = db.Query(query, pubkey); err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		rows.Scan(
			&invoice.Id, &invoice.Pubkey, &invoice.Msats, &invoice.Preimage, &invoice.Hash,
			&invoice.PaymentRequest, &invoice.CreatedAt, &invoice.ExpiresAt, &invoice.ConfirmedAt, &invoice.HeldSince, &invoice.Description, &invoice.Status)
		*invoices = append(*invoices, invoice)
	}
	return nil
}

func (db *DB) ConfirmInvoice(tx *sql.Tx, c context.Context, hash string, confirmedAt time.Time, msatsReceived int) error {
	if _, err := tx.ExecContext(c, "UPDATE invoices SET confirmed_at = $2, msats_received = $3 WHERE hash = $1", hash, confirmedAt, msatsReceived); err != nil {
		return err
	}
	return nil
}
