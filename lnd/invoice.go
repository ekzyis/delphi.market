package lnd

import (
	"context"
	"database/sql"
	"log"
	"time"

	"git.ekzyis.com/ekzyis/delphi.market/db"
	"github.com/lightninglabs/lndclient"
	"github.com/lightningnetwork/lnd/lnrpc/invoicesrpc"
	"github.com/lightningnetwork/lnd/lntypes"
	"github.com/lightningnetwork/lnd/lnwire"
)

func (lnd *LNDClient) CreateInvoice(tx *sql.Tx, ctx context.Context, d *db.DB, pubkey string, msats int64, description string) (*db.Invoice, error) {
	var (
		expiry         time.Duration = time.Hour
		preimage       lntypes.Preimage
		hash           lntypes.Hash
		paymentRequest string
		lnInvoice      *lndclient.Invoice
		dbInvoice      *db.Invoice
		err            error
	)
	if preimage, err = generateNewPreimage(); err != nil {
		return nil, err
	}
	hash = preimage.Hash()
	if paymentRequest, err = lnd.Invoices.AddHoldInvoice(ctx, &invoicesrpc.AddInvoiceData{
		Hash:   &hash,
		Value:  lnwire.MilliSatoshi(msats),
		Expiry: int64(expiry / time.Millisecond),
	}); err != nil {
		return nil, err
	}
	if lnInvoice, err = lnd.Client.LookupInvoice(ctx, hash); err != nil {
		return nil, err
	}
	dbInvoice = &db.Invoice{
		Pubkey:         pubkey,
		Msats:          msats,
		Preimage:       preimage.String(),
		PaymentRequest: paymentRequest,
		Hash:           hash.String(),
		CreatedAt:      lnInvoice.CreationDate,
		ExpiresAt:      lnInvoice.CreationDate.Add(expiry),
		Description:    description,
	}
	if err := d.CreateInvoice(tx, ctx, dbInvoice); err != nil {
		return nil, err
	}
	return dbInvoice, nil
}

func (lnd *LNDClient) CheckInvoice(d *db.DB, hash lntypes.Hash) {
	var (
		pollInterval = 5 * time.Second
		invoice      db.Invoice
		lnInvoice    *lndclient.Invoice
		preimage     lntypes.Preimage
		err          error
	)

	if err = d.FetchInvoice(&db.FetchInvoiceWhere{Hash: hash.String()}, &invoice); err != nil {
		log.Println(err)
		return
	}

	handleLoopError := func(err error) {
		log.Println(err)
		time.Sleep(pollInterval)
	}

	for {
		log.Printf("lookup invoice: hash=%s", hash)
		if lnInvoice, err = lnd.Client.LookupInvoice(context.TODO(), hash); err != nil {
			handleLoopError(err)
			continue
		}
		if time.Now().After(invoice.ExpiresAt) {
			// cancel invoices after expiration if no matching order found yet
			if err = lnd.Invoices.CancelInvoice(context.TODO(), hash); err != nil {
				handleLoopError(err)
				continue
			}
			log.Printf("invoice expired: hash=%s", hash)
			break
		}
		if lnInvoice.AmountPaid == lnInvoice.Amount {
			if preimage, err = lntypes.MakePreimageFromStr(invoice.Preimage); err != nil {
				handleLoopError(err)
				continue
			}
			// TODO settle invoice after matching order was found
			if err = lnd.Invoices.SettleInvoice(context.TODO(), preimage); err != nil {
				handleLoopError(err)
				continue
			}
			if err = d.ConfirmInvoice(hash.String(), time.Now(), int(lnInvoice.AmountPaid)); err != nil {
				handleLoopError(err)
				continue
			}
			log.Printf("invoice confirmed: hash=%s", hash)
			break
		}
		time.Sleep(pollInterval)
	}
}

func (lnd *LNDClient) CheckInvoices(d *db.DB) error {
	var (
		invoices []db.Invoice
		err      error
		hash     lntypes.Hash
	)
	if err = d.FetchInvoices(&db.FetchInvoicesWhere{Unconfirmed: true}, &invoices); err != nil {
		return err
	}
	for _, invoice := range invoices {
		if hash, err = lntypes.MakeHashFromStr(invoice.Hash); err != nil {
			return err
		}
		go lnd.CheckInvoice(d, hash)
	}
	return nil
}
