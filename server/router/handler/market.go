package handler

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"strconv"
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
			b = (float64(cost) / 1000) / math.Log(2)

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
			db       = sc.Db
			ctx      = c.Request().Context()
			id       = c.Param("id")
			quantity = c.QueryParam("q")
			q        int64
			m        = types.Market{}
			u        = types.User{}
			l        = types.LMSR{}
			total    float64
			quote0   = types.MarketQuote{}
			quote1   = types.MarketQuote{}
			err      error
		)

		if quantity == "" {
			q = 1
		} else if q, err = strconv.ParseInt(quantity, 10, 64); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "q must be integer")
		}

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

		total = lmsr.Quote(l.B, l.Q1, l.Q2, int(q))
		quote0 = types.MarketQuote{
			Outcome:    0,
			AvgPrice:   total / float64(q),
			TotalPrice: total,
			Reward:     float64(q) - total,
		}

		total = lmsr.Quote(l.B, l.Q2, l.Q1, int(q))
		quote1 = types.MarketQuote{
			Outcome:    1,
			AvgPrice:   total / float64(q),
			TotalPrice: total,
			Reward:     float64(q) - total,
		}

		return pages.Market(m, quote0, quote1).Render(context.RenderContext(sc, c), c.Response().Writer)
	}
}

func HandleOrder(sc context.Context) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			db             = sc.Db
			lnd            = sc.Lnd
			tx             *sql.Tx
			ctx            = c.Request().Context()
			u              = c.Get("session").(types.User)
			id             = c.Param("id")
			quantity       = c.FormValue("q")
			outcome        = c.FormValue("o")
			q              int64
			o              int64
			m              = types.Market{}
			mU             = types.User{}
			l              = types.LMSR{}
			totalF         float64
			total          int
			hash           lntypes.Hash
			paymentRequest string
			expiry         = int64(60)
			expiresAt      = time.Now().Add(time.Second * time.Duration(expiry))
			invoiceId      int
			invDescription string
			orderId        int
			qr             templ.Component
			err            error
		)

		if quantity == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "q must be given")
		} else if q, err = strconv.ParseInt(quantity, 10, 64); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "q must be integer")
		}

		if outcome == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "o must be given")
		} else if o, err = strconv.ParseInt(outcome, 10, 64); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "o must be integer")
		}
		if o < 0 && o > 1 {
			return echo.NewHTTPError(http.StatusBadRequest, "o must be 0 or 1")
		}

		// TODO: refactor since this uses same queries as function above
		if err = db.QueryRowContext(ctx, ""+
			"SELECT m.id, m.question, m.description, m.created_at, m.end_date, m.lmsr_b, "+
			"u.id, u.name, u.created_at, u.ln_pubkey, u.nostr_pubkey, u.msats "+
			"FROM markets m "+
			"JOIN users u ON m.user_id = u.id "+
			"JOIN invoices i ON m.invoice_id = i.id "+
			"WHERE m.id = $1 AND i.confirmed_at IS NOT NULL", id).Scan(
			&m.Id, &m.Question, &m.Description, &m.CreatedAt, &m.EndDate, &l.B,
			&mU.Id, &mU.Name, &mU.CreatedAt, &mU.LnPubkey, &mU.NostrPubkey, &mU.Msats); err != nil {
			if err == sql.ErrNoRows {
				return echo.NewHTTPError(http.StatusNotFound)
			}
			return err
		}
		m.User = mU

		if err = db.QueryRowContext(ctx, ""+
			"SELECT "+
			"COUNT(o.quantity) FILTER(WHERE o.outcome = 0) AS q1, "+
			"COUNT(o.quantity) FILTER(WHERE o.outcome = 1) AS q2 "+
			"FROM orders o "+
			"JOIN markets m ON o.market_id = m.id "+
			"JOIN invoices i ON o.invoice_id = i.id "+
			"WHERE o.market_id = $1 AND i.confirmed_at IS NOT NULL", id).Scan(
			&l.Q1, &l.Q2); err != nil {
			if err == sql.ErrNoRows {
				return echo.NewHTTPError(http.StatusNotFound)
			}
			return err
		}

		if o == 0 {
			totalF = lmsr.Quote(l.B, l.Q1, l.Q2, int(q))
		} else if o == 1 {
			totalF = lmsr.Quote(l.B, l.Q2, l.Q1, int(q))
		}

		total = int(math.Round(totalF * 1000))

		if tx, err = db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted}); err != nil {
			return err
		}

		if hash, paymentRequest, err = lnd.Client.AddInvoice(ctx,
			&invoicesrpc.AddInvoiceData{
				Value:  lnwire.MilliSatoshi(total),
				Expiry: expiry,
			}); err != nil {
			return err
		}

		if err = tx.QueryRowContext(ctx, ""+
			"INSERT INTO invoices (user_id, msats, hash, bolt11, expires_at) "+
			"VALUES ($1, $2, $3, $4, $5) "+
			"RETURNING id",
			u.Id, total, hash.String(), paymentRequest, expiresAt).Scan(&invoiceId); err != nil {
			return err
		}

		if err = tx.QueryRowContext(ctx, ""+
			"INSERT INTO orders (market_id, user_id, quantity, outcome, invoice_id) "+
			"VALUES ($1, $2, $3, $4, $5) "+
			"RETURNING id",
			id, u.Id, q, o, invoiceId).Scan(&orderId); err != nil {
			return err
		}

		invDescription = fmt.Sprintf("create order %d for market %s", orderId, id)
		if _, err = tx.ExecContext(ctx, ""+
			"UPDATE invoices SET description = $1 WHERE id = $2",
			invDescription, invoiceId); err != nil {
			return err
		}

		if err = tx.Commit(); err != nil {
			return err
		}

		qr = components.Invoice(hash.String(), paymentRequest, total, int(expiry), false, toRedirectUrl(invDescription))

		return components.Modal(qr).Render(context.RenderContext(sc, c), c.Response().Writer)
	}
}
