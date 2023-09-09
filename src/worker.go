package main

import (
	"log"

	"github.com/lightningnetwork/lnd/lntypes"
)

func RunJobs() error {
	var invoices []Invoice
	if err := db.FetchInvoices(&FetchInvoicesWhere{Expired: false}, &invoices); err != nil {
		return err
	}
	for _, inv := range invoices {
		hash, err := lntypes.MakeHashFromStr(inv.PaymentHash)
		if err != nil {
			log.Println(err)
			continue
		}
		go lnd.CheckInvoice(hash)
	}
	return nil
}
