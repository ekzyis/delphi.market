package lnd

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/btcsuite/btcd/btcutil"
)

func (lnd *LNDClient) PayInvoice(tx *sql.Tx, bolt11 string) error {
	maxFeeSats := btcutil.Amount(10)
	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()
	log.Printf("attempting to pay bolt11 %s ...\n", bolt11)
	payChan := lnd.Client.PayInvoice(ctx, bolt11, maxFeeSats, nil)
	res := <-payChan
	if res.Err != nil {
		log.Printf("error paying bolt11: %s -- %s\n", bolt11, res.Err)
		tx.Rollback()
		return res.Err
	}
	log.Printf("successfully paid bolt11: %s\n", bolt11)
	if _, err := tx.ExecContext(ctx, "UPDATE withdrawals SET paid_at = CURRENT_TIMESTAMP WHERE bolt11 = $1", bolt11); err != nil {
		tx.Rollback()
		return err
	}
	return res.Err
}
