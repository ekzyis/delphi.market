package lnd

import (
	"context"
	"database/sql"
	"log"
	"time"

	"git.ekzyis.com/ekzyis/delphi.market/db"
	"github.com/lightninglabs/lndclient"
	"github.com/lightningnetwork/lnd/invoices"
	"github.com/lightningnetwork/lnd/lntypes"
)

type LNDClient struct {
	*lndclient.GrpcLndServices
}

type LNDConfig = lndclient.LndServicesConfig

func New(config *LNDConfig) (*LNDClient, error) {
	var (
		rcpLndServices *lndclient.GrpcLndServices
		lnd            *LNDClient
		err            error
	)
	if rcpLndServices, err = lndclient.NewLndServices(config); err != nil {
		return nil, err
	}
	lnd = &LNDClient{GrpcLndServices: rcpLndServices}
	log.Printf("Connected to %s running LND v%s", config.LndAddress, lnd.Version.Version)
	return lnd, nil
}

func (lnd *LNDClient) CheckInvoices(db *db.DB) {
	var (
		pending bool
		rows    *sql.Rows
		hash    lntypes.Hash
		inv     *lndclient.Invoice
		err     error
	)
	for {
		// fetch all pending invoices
		if rows, err = db.Query("" +
			"SELECT hash, expires_at " +
			"FROM invoices " +
			"WHERE confirmed_at IS NULL AND expires_at > CURRENT_TIMESTAMP"); err != nil {
			log.Printf("error checking invoices: %v", err)
		}

		pending = false

		for rows.Next() {
			pending = true

			var (
				h         string
				expiresAt time.Time
			)
			rows.Scan(&h, &expiresAt)

			if hash, err = lntypes.MakeHashFromStr(h); err != nil {
				log.Printf("error parsing hash: %v", err)
				continue
			}

			if inv, err = lnd.Client.LookupInvoice(context.TODO(), hash); err != nil {
				log.Printf("error looking up invoice: %v", err)
				continue
			}

			if !inv.State.IsFinal() {
				log.Printf("invoice pending: %s %s", h, time.Until(expiresAt))
				continue
			}

			if inv.State == invoices.ContractSettled {
				if _, err = db.Exec(
					"UPDATE invoices SET msats_received = $1, confirmed_at = $2 WHERE hash = $3",
					inv.AmountPaid, inv.SettleDate, h); err != nil {
					log.Printf("error updating invoice %s: %v", h, err)
				}
				log.Printf("invoice confirmed: %s", h)
			} else if inv.State == invoices.ContractCanceled {
				log.Printf("invoice expired: %s", h)
			}
		}

		// poll faster if there are pending invoices
		if pending {
			time.Sleep(1 * time.Second)
		} else {
			time.Sleep(5 * time.Second)
		}
	}
}
