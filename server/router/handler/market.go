package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"git.ekzyis.com/ekzyis/delphi.market/db"
	"git.ekzyis.com/ekzyis/delphi.market/env"
	"git.ekzyis.com/ekzyis/delphi.market/lib"
	"git.ekzyis.com/ekzyis/delphi.market/lnd"
	"github.com/labstack/echo/v4"
	"github.com/lightningnetwork/lnd/lntypes"
)

func HandleMarket(envVars map[string]any) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			marketId int64
			market   db.Market
			shares   []db.Share
			orders   []db.Order
			err      error
			data     map[string]any
		)
		if marketId, err = strconv.ParseInt(c.Param("id"), 10, 64); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Bad Request")
		}
		if err = db.FetchMarket(int(marketId), &market); err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Not Found")
		} else if err != nil {
			return err
		}
		if err = db.FetchShares(market.Id, &shares); err != nil {
			return err
		}
		if err = db.FetchOrders(&db.FetchOrdersWhere{MarketId: market.Id, Confirmed: true}, &orders); err != nil {
			return err
		}
		data = map[string]any{
			"session":     c.Get("session"),
			"Id":          market.Id,
			"Description": market.Description,
			// shares are sorted by description in descending order
			// that's how we know that YES must be the first share
			"YesShare": shares[0],
			"NoShare":  shares[1],
			"Orders":   orders,
		}
		lib.Merge(&data, &envVars)
		return c.Render(http.StatusOK, "market.html", data)
	}
}

func HandlePostOrder(envVars map[string]any) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			marketId string
			u        db.User
			o        db.Order
			invoice  *db.Invoice
			msats    int64
			data     map[string]any
			qr       string
			hash     lntypes.Hash
			err      error
		)
		marketId = c.Param("id")
		// TODO:
		//   [ ] Step 0: If SELL order, check share balance of user
		//   [x] Create HODL invoice
		//   [x] Create (unconfirmed) order
		//   [ ] Find matching orders
		//   [ ] Settle invoice when matching order was found,
		//         else cancel invoice if expired

		// parse body
		if err := c.Bind(&o); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest)
		}
		u = c.Get("session").(db.User)
		o.Pubkey = u.Pubkey
		msats = o.Quantity * o.Price * 1000

		// TODO: if SELL order, check share balance of user

		// Create HODL invoice
		if invoice, err = lnd.CreateInvoice(o.Pubkey, msats); err != nil {
			return err
		}
		// Create QR code to pay HODL invoice
		if qr, err = lib.ToQR(invoice.PaymentRequest); err != nil {
			return err
		}
		if hash, err = lntypes.MakeHashFromStr(invoice.Hash); err != nil {
			return err
		}

		// Start goroutine to poll status and update invoice in background
		go lnd.CheckInvoice(hash)

		// Create (unconfirmed) order
		o.InvoiceId = invoice.Id
		if err := db.CreateOrder(&o); err != nil {
			return err
		}

		// TODO: find matching orders

		data = map[string]any{
			"session":     c.Get("session"),
			"lnurl":       invoice.PaymentRequest,
			"qr":          qr,
			"invoice":     *invoice,
			"redirectURL": fmt.Sprintf("https://%s/market/%s", env.PublicURL, marketId),
		}
		lib.Merge(&data, &envVars)
		return c.Render(http.StatusPaymentRequired, "invoice.html", data)
	}
}
