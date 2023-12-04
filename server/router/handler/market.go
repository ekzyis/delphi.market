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
			"Pubkey":      market.Pubkey,
			"Description": market.Description,
			"SettledAt":   market.SettledAt,
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
			m           db.Market
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

		if err = sc.Db.FetchMarket(s.MarketId, &m); err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, nil)
		} else if err != nil {
			return err
		}

		if m.SettledAt.Valid {
			return c.JSON(http.StatusBadRequest, map[string]string{"reason": "market already settled"})
		}

		description = fmt.Sprintf("%s %d %s shares @ %d sats [market:%d]", strings.ToUpper(o.Side), o.Quantity, s.Description, o.Price, s.MarketId)

		if o.Side == "BUY" {
			// BUY orders require payment
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

func HandleMarketSettlement(sc context.ServerContext) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			marketId int64
			market   db.Market
			s        db.Share
			tx       *sql.Tx
			u        db.User
			query    string
			err      error
		)
		if marketId, err = strconv.ParseInt(c.Param("id"), 10, 64); err != nil {
			return c.JSON(http.StatusBadRequest, nil)
		}

		if err = c.Bind(&s); err != nil || s.Id == "" {
			return c.JSON(http.StatusBadRequest, nil)
		}

		if err = sc.Db.FetchMarket(int(marketId), &market); err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{"reason": "market not found"})
		} else if err != nil {
			return err
		}

		u = c.Get("session").(db.User)

		// only market owner can settle market
		if market.Pubkey != u.Pubkey {
			return c.JSON(http.StatusForbidden, map[string]string{"reason": "not your market"})
		}

		// market already settled?
		if market.SettledAt.Valid {
			return c.JSON(http.StatusBadRequest, map[string]string{"reason": "market already settled"})
		}

		// transaction start
		ctx, cancel := context_.WithTimeout(c.Request().Context(), 10*time.Second)
		defer cancel()
		if tx, err = sc.Db.BeginTx(ctx, nil); err != nil {
			return err
		}
		defer tx.Commit()

		// refund users for pending BUY orders
		query = "" +
			"UPDATE users u SET msats = msats + pending_orders.msats_received FROM ( " +
			"	SELECT o.pubkey, i.msats_received " +
			"   FROM orders o " +
			"	LEFT JOIN invoices i ON i.id = o.invoice_id " +
			"	JOIN shares s ON s.id = o.share_id " +
			"	WHERE s.market_id = $1 " +
			// an order is pending if it wasn't canceled and wasn't matched yet
			// (the o.side = 'BUY' shouldn't be necessary since i.msats_received will be NULL for SELL orders anyway
			//  but added here for clarification anyway)
			"	AND o.side = 'BUY' AND o.deleted_at IS NULL AND o.order_id IS NULL " +
			") AS pending_orders WHERE pending_orders.pubkey = u.pubkey"
		if _, err = tx.ExecContext(ctx, query, marketId); err != nil {
			tx.Rollback()
			return err
		}

		// now cancel pending orders
		query = "" +
			"UPDATE orders o SET deleted_at = CURRENT_TIMESTAMP WHERE id IN ( " +
			// basically same subquery as above
			"  SELECT o.id FROM orders o " +
			"  JOIN shares s ON s.id = o.share_id " +
			// again, orders are pending if they weren't canceled and weren't matched yet
			"  WHERE s.market_id = $1 AND o.deleted_at IS NULL and o.order_id IS NULL " +
			")"
		if _, err = tx.ExecContext(ctx, query, marketId); err != nil {
			tx.Rollback()
			return err
		}

		// payout
		query = "" +
			// * 100 since winning shares expire at 100 sats per share
			// * 1000 to convert sats to msats
			"UPDATE users u SET msats = msats + (user_shares.quantity * 100 * 1000) " +
			"FROM ( " +
			"    SELECT o.pubkey, o.share_id, " +
			"    SUM(CASE WHEN o.side = 'BUY' THEN o.quantity ELSE -o.quantity END) AS quantity " +
			"    FROM orders o " +
			"    LEFT JOIN invoices i ON i.id = o.invoice_id " +
			"    JOIN shares s ON s.id = o.share_id " +
			// only consider uncanceled orders for winning shares
			"    WHERE s.market_id = $1 AND o.deleted_at IS NULL AND s.id = $2 " +
			// BUY orders must be paid and be matched. SELL orders must simply not be canceled to be considered.
			"    AND ( (o.side = 'BUY' AND i.confirmed_at IS NOT NULL AND o.order_id IS NOT NULL) OR o.side = 'SELL' ) " +
			"    GROUP BY o.pubkey, o.share_id " +
			") AS user_shares WHERE user_shares.pubkey = u.pubkey"
		if _, err = tx.ExecContext(ctx, query, marketId, s.Id); err != nil {
			tx.Rollback()
			return err
		}

		if _, err = tx.ExecContext(ctx, "UPDATE markets SET settled_at = CURRENT_TIMESTAMP WHERE id = $1", marketId); err != nil {
			tx.Rollback()
			return err
		}

		if _, err = tx.ExecContext(ctx, "UPDATE shares SET win = (id = $1) WHERE market_id = $2", s.Id, marketId); err != nil {
			tx.Rollback()
			return err
		}

		tx.Commit()

		return c.JSON(http.StatusOK, nil)
	}
}
