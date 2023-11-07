package db

import "time"

func (db *DB) CreateInvoice(invoice *Invoice) error {
	if err := db.QueryRow(""+
		"INSERT INTO invoices(pubkey, msats, preimage, hash, bolt11, created_at, expires_at) "+
		"VALUES($1, $2, $3, $4, $5, $6, $7) "+
		"RETURNING id",
		invoice.Pubkey, invoice.Msats, invoice.Preimage, invoice.Hash, invoice.PaymentRequest, invoice.CreatedAt, invoice.ExpiresAt).Scan(&invoice.Id); err != nil {
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
		query = "SELECT id, pubkey, msats, preimage, hash, bolt11, created_at, expires_at, confirmed_at, held_since FROM invoices "
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
		&invoice.PaymentRequest, &invoice.CreatedAt, &invoice.ExpiresAt, &invoice.ConfirmedAt, &invoice.HeldSince); err != nil {
		return err
	}
	return nil
}

func (db *DB) ConfirmInvoice(hash string, confirmedAt time.Time, msatsReceived int) error {
	if _, err := db.Exec("UPDATE invoices SET confirmed_at = $2, msats_received = $3 WHERE hash = $1", hash, confirmedAt, msatsReceived); err != nil {
		return err
	}
	return nil
}