package main

func RunJobs() error {
	var invoices []Invoice
	if err := db.FetchInvoices(&FetchInvoicesWhere{Expired: false}, &invoices); err != nil {
		return err
	}
	for _, inv := range invoices {
		go lnd.CheckInvoice(inv.PaymentHash)
	}
	return nil
}
