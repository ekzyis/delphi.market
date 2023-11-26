package handler

import (
	context_ "context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

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

func HandleCreateMarket(sc context.ServerContext) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			tx             *sql.Tx
			u              db.User
			m              db.Market
			invoice        *db.Invoice
			msats          int64
			invDescription string
			data           map[string]any
			qr             string
			hash           lntypes.Hash
			err            error
		)
		if err := c.Bind(&m); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest)
		}

		// transaction start
		ctx, cancel := context_.WithTimeout(c.Request().Context(), 5*time.Second)
		defer cancel()
		if tx, err = sc.Db.BeginTx(ctx, nil); err != nil {
			tx.Rollback()
			return err
		}
		defer tx.Commit()

		u = c.Get("session").(db.User)
		msats = 1000
		// TODO: add [market:<id>] for redirect after payment
		invDescription = fmt.Sprintf("create market \"%s\"", m.Description)
		if invoice, err = sc.Lnd.CreateInvoice(tx, ctx, sc.Db, u.Pubkey, msats, invDescription); err != nil {
			tx.Rollback()
			return err
		}
		if qr, err = lib.ToQR(invoice.PaymentRequest); err != nil {
			tx.Rollback()
			return err
		}
		if hash, err = lntypes.MakeHashFromStr(invoice.Hash); err != nil {
			tx.Rollback()
			return err
		}
		m.InvoiceId = invoice.Id
		if err := sc.Db.CreateMarket(tx, ctx, &m); err != nil {
			tx.Rollback()
			return err
		}

		// need to commit before starting to poll invoice status
		tx.Commit()
		go sc.Lnd.CheckInvoice(sc.Db, hash)

		data = map[string]any{
			"id":     invoice.Id,
			"bolt11": invoice.PaymentRequest,
			"amount": msats,
			"qr":     qr,
		}
		return c.JSON(http.StatusPaymentRequired, data)
	}
}

func HandleOrder(sc context.ServerContext) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			tx          *sql.Tx
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
		//   [ ] If SELL order, check share balance of user
		//   [x] Create (unconfirmed) order
		//   [x] Create invoice
		//   [ ] Find matching orders
		//   [ ] show invoice to user

		// parse body
		if err := c.Bind(&o); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest)
		}
		u = c.Get("session").(db.User)
		o.Pubkey = u.Pubkey
		msats = o.Quantity * o.Price * 1000

		// transaction start
		ctx, cancel := context_.WithTimeout(c.Request().Context(), 5*time.Second)
		defer cancel()
		if tx, err = sc.Db.BeginTx(ctx, nil); err != nil {
			tx.Rollback()
			return err
		}
		defer tx.Commit()

		if err = sc.Db.FetchShare(tx, ctx, o.ShareId, &s); err != nil {
			tx.Rollback()
			return err
		}
		description = fmt.Sprintf("%s %d %s shares @ %d sats [market:%d]", strings.ToUpper(o.Side), o.Quantity, s.Description, o.Price, s.MarketId)

		// TODO: if SELL order, check share balance of user

		// Create HODL invoice
		if invoice, err = sc.Lnd.CreateInvoice(tx, ctx, sc.Db, o.Pubkey, msats, description); err != nil {
			tx.Rollback()
			return err
		}
		// Create QR code to pay HODL invoice
		if qr, err = lib.ToQR(invoice.PaymentRequest); err != nil {
			tx.Rollback()
			return err
		}
		if hash, err = lntypes.MakeHashFromStr(invoice.Hash); err != nil {
			tx.Rollback()
			return err
		}

		// Create (unconfirmed) order
		o.InvoiceId = invoice.Id
		if err := sc.Db.CreateOrder(tx, ctx, &o); err != nil {
			tx.Rollback()
			return err
		}

		// need to commit before startign to poll invoice status
		tx.Commit()
		go sc.Lnd.CheckInvoice(sc.Db, hash)

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

func HandleOrders(sc context.ServerContext) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			u      db.User
			orders []db.Order
			err    error
		)
		u = c.Get("session").(db.User)
		if err = sc.Db.FetchUserOrders(u.Pubkey, &orders); err != nil {
			return err
		}
		return c.JSON(http.StatusOK, orders)
	}
}
