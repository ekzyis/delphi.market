package handler

import (
	"database/sql"

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
			err     error
			markets []types.Market
		)

		if rows, err = db.QueryContext(ctx, ""+
			"SELECT m.id, m.question, m.description, m.created_at, m.end_date, "+
			"u.id, u.name, u.created_at, u.ln_pubkey, u.nostr_pubkey, u.msats "+
			"FROM markets m "+
			"JOIN users u ON m.user_id = u.id "+
			"JOIN invoices i ON m.invoice_id = i.id "+
			"WHERE i.confirmed_at IS NOT NULL"); err != nil {
			return err
		}

		for rows.Next() {
			var m types.Market
			var u types.User
			if err = rows.Scan(
				&m.Id, &m.Question, &m.Description, &m.CreatedAt, &m.EndDate,
				&u.Id, &u.Name, &u.CreatedAt, &u.LnPubkey, &u.NostrPubkey, &u.Msats); err != nil {
				return err
			}
			m.User = u
			markets = append(markets, m)
		}

		return pages.Index(markets).Render(context.RenderContext(sc, c), c.Response().Writer)
	}
}
