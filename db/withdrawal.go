package db

import (
	"context"
	"database/sql"
)

func (db *DB) CreateWithdrawal(tx *sql.Tx, c context.Context, w *Withdrawal) error {
	if _, err := tx.ExecContext(c, "INSERT INTO withdrawals(pubkey, bolt11) VALUES ($1, $2)", w.Pubkey, w.Bolt11); err != nil {
		return err
	}
	return nil
}
