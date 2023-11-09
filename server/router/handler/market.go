package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"git.ekzyis.com/ekzyis/delphi.market/db"
	"git.ekzyis.com/ekzyis/delphi.market/lib"
	"git.ekzyis.com/ekzyis/delphi.market/server/router/context"
	"github.com/labstack/echo/v4"
	"github.com/lightningnetwork/lnd/lntypes"
)

func HandleMarket(sc context.ServerContext) echo.HandlerFunc {
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
		if err = sc.Db.FetchMarket(int(marketId), &market); err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Not Found")
		} else if err != nil {
			return err
		}
		if err = sc.Db.FetchShares(market.Id, &shares); err != nil {
			return err
		}
		if err = sc.Db.FetchOrders(&db.FetchOrdersWhere{MarketId: market.Id, Confirmed: true}, &orders); err != nil {
			return err
		}
		data = map[string]any{
			"Id":          market.Id,
			"Description": market.Description,
			"Shares":      shares,
		}
		return c.JSON(http.StatusOK, data)
	}
}

func HandleOrder(sc context.ServerContext) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			u           db.User
			o           db.Order
			s           db.Share
			invoice     *db.Invoice
			msats       int64
			description string
			data        map[string]any
			qr          string
			hash        lntypes.Hash
			err         error
		)
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
		if err = sc.Db.FetchShare(o.ShareId, &s); err != nil {
			return err
		}
		description = fmt.Sprintf("%s %d %s shares @ %d [market:%d]", strings.ToUpper(o.Side), o.Quantity, s.Description, o.Price, s.MarketId)

		// TODO: if SELL order, check share balance of user

		// Create HODL invoice
		if invoice, err = sc.Lnd.CreateInvoice(sc.Db, o.Pubkey, msats, description); err != nil {
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
		go sc.Lnd.CheckInvoice(sc.Db, hash)

		// Create (unconfirmed) order
		o.InvoiceId = invoice.Id
		if err := sc.Db.CreateOrder(&o); err != nil {
			return err
		}

		// TODO: find matching orders

		data = map[string]any{
			"id":     invoice.Id,
			"bolt11": invoice.PaymentRequest,
			"amount": msats,
			"qr":     qr,
		}
		return c.JSON(http.StatusPaymentRequired, data)
	}
}
