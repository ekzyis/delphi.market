package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

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
			cost           = lnwire.MilliSatoshi(1000e3)
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
			"INSERT INTO markets (question, description, end_date, user_id, invoice_id) "+
			"VALUES ($1, $2, $3, $4, $5) "+
			"RETURNING id",
			question, description, endDate, u.Id, invoiceId).Scan(&marketId); err != nil {
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
			err error
		)

		if err = db.QueryRowContext(ctx, ""+
			"SELECT m.id, m.question, m.description, m.created_at, m.end_date, "+
			"u.id, u.name, u.created_at, u.ln_pubkey, u.nostr_pubkey, u.msats "+
			"FROM markets m "+
			"JOIN users u ON m.user_id = u.id "+
			"JOIN invoices i ON m.invoice_id = i.id "+
			"WHERE m.id = $1 AND i.confirmed_at IS NOT NULL", id).Scan(
			&m.Id, &m.Question, &m.Description, &m.CreatedAt, &m.EndDate,
			&u.Id, &u.Name, &u.CreatedAt, &u.LnPubkey, &u.NostrPubkey, &u.Msats); err != nil {
			if err == sql.ErrNoRows {
				return echo.NewHTTPError(http.StatusNotFound)
			}
			return err
		}

		m.User = u

		return pages.Market(m).Render(context.RenderContext(sc, c), c.Response().Writer)
	}
}
