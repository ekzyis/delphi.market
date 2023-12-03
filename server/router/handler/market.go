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
			u        db.User
			tx       *sql.Tx
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
		if session := c.Get("session"); session != nil {
			u = session.(db.User)
			ctx, cancel := context_.WithTimeout(context_.TODO(), 10*time.Second)
			defer cancel()
			if tx, err = sc.Db.BeginTx(ctx, nil); err != nil {
				return err
			}
			defer tx.Commit()
			uBalance := make(map[string]any)
			if err = sc.Db.FetchUserBalance(tx, ctx, int(marketId), u.Pubkey, &uBalance); err != nil {
				tx.Rollback()
				return err
			}
			lib.Merge(&data, &map[string]any{"user": uBalance})
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
			return err
		}
		defer tx.Commit()

		u = c.Get("session").(db.User)
		m.Pubkey = u.Pubkey
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

func HandleMarketOrders(sc context.ServerContext) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			marketId int64
			orders   []db.Order
			err      error
		)
		if marketId, err = strconv.ParseInt(c.Param("id"), 10, 64); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Bad Request")
		}
		if err = sc.Db.FetchMarketOrders(marketId, &orders); err != nil {
			return err
		}
		return c.JSON(http.StatusOK, orders)
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
			return err
		}
		defer tx.Commit()

		if err = sc.Db.FetchShare(tx, ctx, o.ShareId, &s); err != nil {
			tx.Rollback()
			return err
		}
		description = fmt.Sprintf("%s %d %s shares @ %d sats [market:%d]", strings.ToUpper(o.Side), o.Quantity, s.Description, o.Price, s.MarketId)

		if o.Side == "BUY" {
			// === Create invoice ===
			// We do this for BUY and SELL orders such that we can continue to use `invoice.confirmed_at IS NOT NULL`
			// to check for confirmed orders
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
			o.InvoiceId.String = invoice.Id
			if err := sc.Db.CreateOrder(tx, ctx, &o); err != nil {
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

		// sell order: check user balance
		balance := make(map[string]any)
		if err = sc.Db.FetchUserBalance(tx, ctx, s.MarketId, o.Pubkey, &balance); err != nil {
			return err
		}
		if balance[s.Description].(int) < int(o.Quantity) {
			tx.Rollback()
			return c.JSON(http.StatusBadRequest, nil)
		}
		// SELL orders don't require payment by user
		if err := sc.Db.CreateOrder(tx, ctx, &o); err != nil {
			tx.Rollback()
			return err
		}
		tx.Commit()
		return c.JSON(http.StatusCreated, nil)
	}
}

func HandleDeleteOrder(sc context.ServerContext) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			orderId string
			tx      *sql.Tx
			u       db.User
			o       db.Order
			msats   int64
			err     error
		)

		if orderId = c.Param("id"); orderId == "" {
			return echo.NewHTTPError(http.StatusBadRequest)
		}
		u = c.Get("session").(db.User)

		// transaction start
		ctx, cancel := context_.WithTimeout(c.Request().Context(), 5*time.Second)
		defer cancel()
		if tx, err = sc.Db.BeginTx(ctx, nil); err != nil {
			return err
		}
		defer tx.Commit()

		if err = sc.Db.FetchOrder(tx, ctx, orderId, &o); err != nil {
			tx.Rollback()
			return err
		}

		if u.Pubkey != o.Pubkey {
			// order does not belong to user
			tx.Rollback()
			return echo.NewHTTPError(http.StatusForbidden)
		}

		if o.OrderId.Valid {
			// order already settled
			tx.Rollback()
			return echo.NewHTTPError(http.StatusBadRequest)
		}

		if o.DeletedAt.Valid {
			// order already deleted
			tx.Rollback()
			return echo.NewHTTPError(http.StatusBadRequest)
		}

		if o.Invoice.ConfirmedAt.Valid {
			// order already paid: we need to move paid sats to user balance before deleting the order
			// TODO update order and session on client
			msats = o.Invoice.MsatsReceived
			if res, err := tx.ExecContext(ctx, "UPDATE users SET msats = msats + $1 WHERE pubkey = $2", msats, u.Pubkey); err != nil {
				tx.Rollback()
				return err
			} else {
				// make sure exactly one row was affected
				if rowsAffected, err := res.RowsAffected(); err != nil {
					tx.Rollback()
					return err
				} else if rowsAffected != 1 {
					tx.Rollback()
					return echo.NewHTTPError(http.StatusInternalServerError)
				}
			}
		}

		if _, err = tx.ExecContext(ctx, "UPDATE orders SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1", o.Id); err != nil {
			tx.Rollback()
			return err
		}

		tx.Commit()

		return c.JSON(http.StatusOK, nil)
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

func HandleMarketStats(sc context.ServerContext) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			marketId int64
			stats    db.MarketStats
			err      error
		)
		if marketId, err = strconv.ParseInt(c.Param("id"), 10, 64); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Bad Request")
		}
		if err = sc.Db.FetchMarketStats(marketId, &stats); err != nil {
			return err
		}
		return c.JSON(http.StatusOK, stats)
	}
}
