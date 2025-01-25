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

func HandleNew(sc context.Context) echo.HandlerFunc {
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
			db  = sc.Db
			ctx = c.Request().Context()

			// session user
			u = types.User{}

			// market id
			id = c.Param("id")

			// quantity of shares user entered into form
			quantity = c.QueryParam("q")

			// quantity as number
			q int64

			// current market
			m = types.Market{}

			// market founder
			mU = types.User{}

			// market LMSR data
			l = types.LMSR{}

			// total price for current quantity of shares in sats
			total float64

			// market quotes
			quoteNo  = types.MarketQuote{}
			quoteYes = types.MarketQuote{}

			// how many shares the user already holds
			uQuantityNo  int
			uQuantityYes int
			rows         *sql.Rows

			// chart data
			lineNo  []types.Point
			lineYes []types.Point

			err error
		)

		if c.Get("session") != nil {
			u = c.Get("session").(types.User)
		} else {
			// unauthenticated user
			u.Id = -1
		}

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
			&mU.Id, &mU.Name, &mU.CreatedAt, &mU.LnPubkey, &mU.NostrPubkey, &mU.Msats); err != nil {
			if err == sql.ErrNoRows {
				return echo.NewHTTPError(http.StatusNotFound)
			}
			return err
		}
		m.User = mU

		if err = db.QueryRowContext(ctx, ""+
			"SELECT "+
			"COALESCE(SUM(i.msats_received), 0) / 1000 AS volume, "+
			"COALESCE(SUM(o.quantity) FILTER(WHERE o.outcome = 0), 0) AS q1, "+
			"COALESCE(SUM(o.quantity) FILTER(WHERE o.outcome = 1), 0) AS q2, "+
			"COALESCE(SUM(o.quantity) FILTER(WHERE o.outcome = 0 AND o.user_id = $2), 0) AS uq1, "+
			"COALESCE(SUM(o.quantity) FILTER(WHERE o.outcome = 1 AND o.user_id = $2), 0) AS uq2 "+
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
			"WHERE o.market_id = $1 AND i.confirmed_at IS NOT NULL", id, u.Id).
			Scan(&m.Volume, &l.Q1, &l.Q2, &uQuantityNo, &uQuantityYes); err != nil {
			if err == sql.ErrNoRows {
				return echo.NewHTTPError(http.StatusNotFound)
			}
			return err
		}

		m.Pyes = lmsr.Quote(l.B, l.Q2, l.Q1, 1)

		if rows, err = db.QueryContext(ctx, ""+
			"SELECT created_at, quote(b, q0, q1, 1) AS p0, quote(b, q1, q0, 1) AS p1 "+
			"FROM ( "+
			"  SELECT "+
			"  m.lmsr_b AS b, o.created_at, "+
			"  COALESCE(SUM(quantity) FILTER(WHERE o.outcome = 0) OVER (ORDER BY o.created_at ASC), 0) AS q0, "+
			"  COALESCE(sum(quantity) filter(where o.outcome = 1) over (order by o.created_at ASC), 0) AS q1 "+
			"  FROM markets m "+
			"  JOIN orders o ON o.market_id = m.id "+
			"  JOIN invoices i ON i.id = o.invoice_id "+
			"  WHERE m.id = $1 AND i.confirmed_at IS NOT NULL "+
			") AS o "+
			"UNION "+
			"SELECT m.created_at, quote(m.lmsr_b, 0, 0, 1) AS p0, quote(m.lmsr_b, 0, 0, 1) AS p1 "+
			"FROM markets m "+
			"ORDER BY created_at", id); err != nil {
			return err
		}

		for rows.Next() {
			var (
				createdAt time.Time
				_p0       float64
				_p1       float64
			)
			if err = rows.Scan(&createdAt, &_p0, &_p1); err != nil {
				return err
			}
			lineNo = append(lineNo, types.Point{X: createdAt, Y: _p0})
			lineYes = append(lineYes, types.Point{X: createdAt, Y: _p1})
		}

		total = lmsr.Quote(l.B, l.Q1, l.Q2, int(q))
		quoteNo = types.MarketQuote{
			Outcome:    0,
			AvgPrice:   total / float64(q),
			TotalPrice: total,
			Reward:     float64(q) - total,
		}

		total = lmsr.Quote(l.B, l.Q2, l.Q1, int(q))
		quoteYes = types.MarketQuote{
			Outcome:    1,
			AvgPrice:   total / float64(q),
			TotalPrice: total,
			Reward:     float64(q) - total,
		}

		return pages.Market(
			m,
			lineNo, lineYes,
			quoteNo, quoteYes,
			uQuantityNo, uQuantityYes).Render(context.RenderContext(sc, c), c.Response().Writer)
	}
}

func HandleOrder(sc context.Context) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			db  = sc.Db
			lnd = sc.Lnd
			tx  *sql.Tx
			ctx = c.Request().Context()
			u   = c.Get("session").(types.User)

			// market id
			id = c.Param("id")

			// how many shares user wants to buy
			quantity = c.FormValue("q")
			// quantity as number
			q int64

			// on which outcome user wants to bet
			outcome = c.FormValue("o")
			// outcome as id
			o int64

			// selected market
			m = types.Market{}

			// market founder
			mU = types.User{}

			// market LMSR data
			l = types.LMSR{}

			// total price as returned by LMSR for given quantity
			totalF float64
			// total rounded to msats
			total int

			// invoice data
			hash           lntypes.Hash
			paymentRequest string
			expiry         = int64(60)
			expiresAt      = time.Now().Add(time.Second * time.Duration(expiry))
			invoiceId      int
			invDescription string

			// id of created order
			orderId int

			// QR component during render
			qr templ.Component

			err error
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
			"COALESCE(SUM(o.quantity) FILTER(WHERE o.outcome = 0), 0) AS q1, "+
			"COALESCE(SUM(o.quantity) FILTER(WHERE o.outcome = 1), 0) AS q2 "+
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
