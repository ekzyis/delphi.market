package lnd

import (
	"context"
	"log"
	"time"

	"git.ekzyis.com/ekzyis/delphi.market/db"
	"github.com/lightninglabs/lndclient"
	"github.com/lightningnetwork/lnd/lnrpc/invoicesrpc"
	"github.com/lightningnetwork/lnd/lntypes"
	"github.com/lightningnetwork/lnd/lnwire"
)

func CreateInvoice(pubkey string, msats int64) (*db.Invoice, error) {
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
	if paymentRequest, err = lnd.Invoices.AddHoldInvoice(context.TODO(), &invoicesrpc.AddInvoiceData{
		Hash:   &hash,
		Value:  lnwire.MilliSatoshi(msats),
		Expiry: int64(expiry / time.Millisecond),
	}); err != nil {
		return nil, err
	}
	if lnInvoice, err = lnd.Client.LookupInvoice(context.TODO(), hash); err != nil {
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
	}
	if err := db.CreateInvoice(dbInvoice); err != nil {
		return nil, err
	}
	return dbInvoice, nil
}

func CheckInvoice(hash lntypes.Hash) {
	var (
		pollInterval = 5 * time.Second
		invoice      db.Invoice
		lnInvoice    *lndclient.Invoice
		preimage     lntypes.Preimage
		err          error
	)

	if !Enabled {
		log.Printf("LND disabled, skipping checking invoice: hash=%s", hash)
		return
	}

	if err = db.FetchInvoice(&db.FetchInvoiceWhere{Hash: hash.String()}, &invoice); err != nil {
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
		if lnInvoice.AmountPaid > 0 {
			if preimage, err = lntypes.MakePreimageFromStr(invoice.Preimage); err != nil {
				handleLoopError(err)
				continue
			}
			// TODO settle invoice after matching order was found
			if err = lnd.Invoices.SettleInvoice(context.TODO(), preimage); err != nil {
				handleLoopError(err)
				continue
			}
			if err = db.ConfirmInvoice(hash.String(), time.Now(), int(lnInvoice.AmountPaid)); err != nil {
				handleLoopError(err)
				continue
			}
			log.Printf("invoice confirmed: hash=%s", hash)
			break
		}
		time.Sleep(pollInterval)
	}
}
