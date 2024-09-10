package handler

import (
	"database/sql"

	"git.ekzyis.com/ekzyis/delphi.market/lib/lmsr"
	"git.ekzyis.com/ekzyis/delphi.market/server/router/context"
	"git.ekzyis.com/ekzyis/delphi.market/server/router/pages"
	"git.ekzyis.com/ekzyis/delphi.market/types"
	"github.com/labstack/echo/v4"
)

func HandleIndex(sc context.Context) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			db      = sc.Db
			ctx     = c.Request().Context()
			rows    *sql.Rows
			markets []types.Market
			err     error
		)

		if rows, err = db.QueryContext(ctx, ""+
			"WITH lmsr AS ("+
			"  SELECT "+
			"    o.market_id, "+
			"    COALESCE(SUM(o.quantity) FILTER(WHERE o.outcome = 0), 0) AS q1, "+
			"    COALESCE(SUM(o.quantity) FILTER(WHERE o.outcome = 1), 0) AS q2, "+
			"    COALESCE(SUM(i.msats_received), 0) / 1000 AS volume "+
			"  FROM orders o "+
			"  JOIN invoices i ON o.invoice_id = i.id "+
			"  WHERE i.confirmed_at IS NOT NULL "+
			"  GROUP BY o.market_id"+
			")"+
			"SELECT m.id, m.question, m.description, m.created_at, m.end_date, "+
			"m.lmsr_b, l.q1, l.q2, l.volume, "+
			"u.id, u.name, u.created_at, u.ln_pubkey, u.nostr_pubkey, u.msats "+
			"FROM markets m "+
			"JOIN users u ON m.user_id = u.id "+
			"JOIN invoices i ON m.invoice_id = i.id "+
			"JOIN lmsr l ON m.id = l.market_id "+
			"WHERE i.confirmed_at IS NOT NULL"); err != nil {
			return err
		}

		for rows.Next() {
			var m types.Market
			var u types.User
			var l types.LMSR
			if err = rows.Scan(
				&m.Id, &m.Question, &m.Description, &m.CreatedAt, &m.EndDate,
				&l.B, &l.Q1, &l.Q2, &m.Volume,
				&u.Id, &u.Name, &u.CreatedAt, &u.LnPubkey, &u.NostrPubkey, &u.Msats); err != nil {
				return err
			}
			m.User = u
			m.Pyes = lmsr.Quote(l.B, l.Q2, l.Q1, 1)

			markets = append(markets, m)
		}

		return pages.Index(markets).Render(context.RenderContext(sc, c), c.Response().Writer)
	}
}
