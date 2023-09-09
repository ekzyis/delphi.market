package main

import (
	"database/sql"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"
)

type Market struct {
	Id          int
	Description string
	Active      bool
}

type Share struct {
	Id          string
	MarketId    int
	Description string
}

type Order struct {
	Session
	Id        string
	ShareId   string `form:"share_id"`
	Side      string `form:"side"`
	Price     int    `form:"price"`
	Quantity  int    `form:"quantity"`
	OrderId   string
	InvoiceId string
}

type Invoice struct {
	Session
	Id             string
	Msats          int
	ReceivedMsats  int
	Preimage       string
	PaymentRequest string
	PaymentHash    string
	CreatedAt      time.Time
	ExpiresAt      time.Time
	ConfirmedAt    null.Time
	HeldSince      null.Time
}

func costFunction(b float64, q1 float64, q2 float64) float64 {
	// reference: http://blog.oddhead.com/2006/10/30/implementing-hansons-market-maker/
	return b * math.Log(math.Pow(math.E, q1/b)+math.Pow(math.E, q2/b))
}

// logarithmic market scoring rule (LMSR) market maker from Robin Hanson:
// https://mason.gmu.edu/~rhanson/mktscore.pdf1
func BinaryLMSR(invariant int, funding int, q1 int, q2 int, dq1 int) float64 {
	b := float64(funding)
	fq1 := float64(q1)
	fq2 := float64(q2)
	fdq1 := float64(dq1)
	return costFunction(b, fq1+fdq1, fq2) - costFunction(b, fq1, fq2)
}

func order(c echo.Context) error {
	marketId := c.Param("id")
	// (TBD) Step 0: If SELL order, check share balance of user
	// (TBD) Step 1: respond with HODL invoice
	o := new(Order)
	if err := c.Bind(o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "bad request")
	}
	session := c.Get("session").(Session)
	o.Pubkey = session.Pubkey
	amount := o.Quantity * o.Price
	invoice, err := lnd.CreateInvoice(session.Pubkey, amount*1000)
	if err != nil {
		return err
	}
	o.InvoiceId = invoice.Id
	if err := db.CreateOrder(o); err != nil {
		return err
	}
	qr, err := ToQR(invoice.PaymentRequest)
	if err != nil {
		return err
	}
	go lnd.CheckInvoice(invoice.PaymentHash)
	data := map[string]any{
		"session":              c.Get("session"),
		"ENV":                  ENV,
		"lnurl":                invoice.PaymentRequest,
		"qr":                   qr,
		"Invoice":              invoice,
		"RedirectAfterPayment": true,
		"PUBLIC_URL":           PUBLIC_URL,
		"MarketId":             marketId,
	}
	return c.Render(http.StatusPaymentRequired, "invoice.html", data)
	// Step 2: After payment, confirm order if no matching order was found
	// if err := db.CreateOrder(o); err != nil {
	// 	if strings.Contains(err.Error(), "violates check constraint") {
	// 		return echo.NewHTTPError(http.StatusBadRequest, "Bad Request")
	// 	}
	// 	return err
	// }
	// (TBD) Step 3:
	//    Settle hodl invoice when matching order was found
	//    else cancel hodl invoice if expired
	// ...
	// return market(c)
}

func market(c echo.Context) error {
	marketId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad Request")
	}

	var market Market
	if err = db.FetchMarket(int(marketId), &market); err == sql.ErrNoRows {
		return echo.NewHTTPError(http.StatusNotFound, "Not Found")
	} else if err != nil {
		return err
	}

	var shares []Share
	if err = db.FetchShares(market.Id, &shares); err != nil {
		return err
	}
	var orders []Order
	if err = db.FetchOrders(market.Id, &orders); err != nil {
		return err
	}
	data := map[string]any{
		"session":     c.Get("session"),
		"ENV":         ENV,
		"Id":          market.Id,
		"Description": market.Description,
		// shares are sorted by description in descending order
		// that's how we know that YES must be the first share
		"YesShare": shares[0],
		"NoShare":  shares[1],
		"Orders":   orders,
	}
	return c.Render(http.StatusOK, "market.html", data)
}
