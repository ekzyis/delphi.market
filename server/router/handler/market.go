package handler

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"time"

	"git.ekzyis.com/ekzyis/delphi.market/lib/lmsr"
	"git.ekzyis.com/ekzyis/delphi.market/server/router/context"
	"git.ekzyis.com/ekzyis/delphi.market/server/router/pages"
	"git.ekzyis.com/ekzyis/delphi.market/server/router/pages/components"
	"git.ekzyis.com/ekzyis/delphi.market/types"
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/lightningnetwork/lnd/lnrpc/invoicesrpc"
	"github.com/lightningnetwork/lnd/lntypes"
	"github.com/lightningnetwork/lnd/lnwire"
)

func HandleCreate(sc context.Context) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			db             = sc.Db
			lnd            = sc.Lnd
			tx             *sql.Tx
			ctx            = c.Request().Context()
			u              = c.Get("session").(types.User)
			question       = c.FormValue("question")
			description    = c.FormValue("description")
			endDate        = c.FormValue("end_date")
			hash           lntypes.Hash
			paymentRequest string
			cost           = lnwire.MilliSatoshi(10_000e3) // creating a market costs 10k sats

			// The cost is used to fund the market.
			// Maximum possible amount of money the market maker can lose is b*ln2.
			// This means if we can only payout as many sats as we paid for the market,
			// we need to solve for b: b = cost / ln2
			b = float64(cost) / math.Log(2)

			expiry         = int64(600)
			expiresAt      = time.Now().Add(time.Second * time.Duration(expiry))
			invoiceId      int
			marketId       int
			invDescription string
			qr             templ.Component
			err            error
		)
		// TODO: validation

		if tx, err = db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted}); err != nil {
			return err
		}

		if hash, paymentRequest, err = lnd.Client.AddInvoice(ctx,
			&invoicesrpc.AddInvoiceData{
				Value:  cost,
				Expiry: expiry,
			}); err != nil {
			return err
		}

		if err = tx.QueryRowContext(ctx, ""+
			"INSERT INTO invoices (user_id, msats, hash, bolt11, expires_at) "+
			"VALUES ($1, $2, $3, $4, $5) "+
			"RETURNING id",
			u.Id, cost, hash.String(), paymentRequest, expiresAt).Scan(&invoiceId); err != nil {
			return err
		}

		if err = tx.QueryRowContext(ctx, ""+
			"INSERT INTO markets (question, description, end_date, user_id, invoice_id, lmsr_b) "+
			"VALUES ($1, $2, $3, $4, $5, $6) "+
			"RETURNING id",
			question, description, endDate, u.Id, invoiceId, b).Scan(&marketId); err != nil {
			return err
		}

		invDescription = fmt.Sprintf("create market %d", marketId)
		if _, err = tx.ExecContext(ctx, ""+
			"UPDATE invoices SET description = $1 WHERE id = $2",
			invDescription, invoiceId); err != nil {
			return err
		}

		if err = tx.Commit(); err != nil {
			return err
		}

		qr = components.Invoice(hash.String(), paymentRequest, int(cost), int(expiry), false, toRedirectUrl(invDescription))

		return components.Modal(qr).Render(context.RenderContext(sc, c), c.Response().Writer)
	}
}

func HandleMarket(sc context.Context) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			db  = sc.Db
			ctx = c.Request().Context()
			id  = c.Param("id")
			m   = types.Market{}
			u   = types.User{}
			l   = types.LSMR{}
			err error
		)

		if err = db.QueryRowContext(ctx, ""+
			"SELECT m.id, m.question, m.description, m.created_at, m.end_date, m.lmsr_b, "+
			"u.id, u.name, u.created_at, u.ln_pubkey, u.nostr_pubkey, u.msats "+
			"FROM markets m "+
			"JOIN users u ON m.user_id = u.id "+
			"JOIN invoices i ON m.invoice_id = i.id "+
			"WHERE m.id = $1 AND i.confirmed_at IS NOT NULL", id).Scan(
			&m.Id, &m.Question, &m.Description, &m.CreatedAt, &m.EndDate, &l.B,
			&u.Id, &u.Name, &u.CreatedAt, &u.LnPubkey, &u.NostrPubkey, &u.Msats); err != nil {
			if err == sql.ErrNoRows {
				return echo.NewHTTPError(http.StatusNotFound)
			}
			return err
		}
		m.User = u

		if err = db.QueryRowContext(ctx, ""+
			"SELECT "+
			"COUNT(o.quantity) FILTER(WHERE o.outcome = 0) AS q1, "+
			"COUNT(o.quantity) FILTER(WHERE o.outcome = 1) AS q2 "+
			"FROM orders o "+
			"JOIN markets m ON o.market_id = m.id "+
			"JOIN invoices i ON o.invoice_id = i.id "+
			// QUESTION: Should unpaid orders contribute to quantity or not?
			//
			// The answer is relevant for concurrent orders:
			// If they do, one can artificially increase the price for others
			// by creating a lot of pending orders.
			// If they don't, one can buy infinite amount of shares at the same price
			// by creating a lot of small but concurrent orders.
			// I think this means that pending order must be scoped to a user
			// but this isn't sybil resistant.
			//
			// For now, we will ignore pending orders.
			"WHERE o.market_id = $1 AND i.confirmed_at IS NOT NULL", id).Scan(
			&l.Q1, &l.Q2); err != nil {
			if err == sql.ErrNoRows {
				return echo.NewHTTPError(http.StatusNotFound)
			}
			return err
		}

		return pages.Market(m, types.MarketP{
			Pyes: lmsr.Price(l.B, l.Q2, l.Q1),
			Pno:  lmsr.Price(l.B, l.Q1, l.Q2), // prices

		}).Render(context.RenderContext(sc, c), c.Response().Writer)
	}
}

func GetPrice(sc context.Context) echo.HandlerFunc {
	return func(c echo.Context) error {

		var (
			db  = sc.Db
			ctx = c.Request().Context()
			id  = c.Param("id")
			m   = types.Market{}
			u   = types.User{}
			err error
		)

		if err = db.QueryRowContext(ctx, ""+
			"SELECT m.id, m.question, m.description, m.created_at, m.end_date, "+
			"u.id, u.name, u.created_at, u.ln_pubkey, u.nostr_pubkey, msats "+
			"FROM markets m JOIN users u ON m.user_id = u.id "+
			"WHERE m.id = $1", id).Scan(
			&m.Id, &m.Question, &m.Description, &m.CreatedAt, &m.EndDate,
			&u.Id, &u.Name, &u.CreatedAt, &u.LnPubkey, &u.NostrPubkey, &u.Msats); err != nil {
			if err == sql.ErrNoRows {
				return echo.NewHTTPError(http.StatusNotFound)
			}
			return err
		}

		m.User = u

		return nil
	}
}
