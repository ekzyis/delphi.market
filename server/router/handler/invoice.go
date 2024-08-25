package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"git.ekzyis.com/ekzyis/delphi.market/server/router/context"
	"git.ekzyis.com/ekzyis/delphi.market/server/router/pages/components"
	"git.ekzyis.com/ekzyis/delphi.market/types"
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func HandleInvoice(sc context.Context) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			db          = sc.Db
			ctx         = c.Request().Context()
			hash        = c.Param("hash")
			u           = c.Get("session").(types.User)
			inv         = types.Invoice{}
			expiresIn   int
			paid        bool
			redirectUrl templ.SafeURL
			qr          templ.Component
			err         error
		)

		if err = db.QueryRowContext(ctx, ""+
			"SELECT user_id, msats, COALESCE(msats_received, 0), expires_at, confirmed_at, bolt11, COALESCE(description, '') "+
			"FROM invoices "+
			"WHERE hash = $1", hash).
			Scan(&inv.UserId, &inv.Msats, &inv.MsatsReceived, &inv.ExpiresAt, &inv.ConfirmedAt, &inv.Bolt11, &inv.Description); err != nil {
			if err == sql.ErrNoRows {
				return echo.NewHTTPError(http.StatusNotFound)
			}
			c.Logger().Error(err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		if u.Id != inv.UserId {
			return echo.NewHTTPError(http.StatusNotFound)
		}

		expiresIn = int(time.Until(inv.ExpiresAt).Seconds())
		paid = inv.MsatsReceived >= inv.Msats
		redirectUrl = toRedirectUrl(inv.Description)

		qr = components.Invoice(hash, inv.Bolt11, int(inv.Msats), expiresIn, paid, redirectUrl)

		return components.Modal(qr).Render(context.RenderContext(sc, c), c.Response().Writer)
	}
}

var (
	marketRegexp = regexp.MustCompile("^create market (?P<id>[0-9]+)$")
)

func toRedirectUrl(description string) templ.SafeURL {
	var m []string
	if m = marketRegexp.FindStringSubmatch(description); m != nil {
		marketId := m[marketRegexp.SubexpIndex("id")]
		return templ.SafeURL(fmt.Sprintf("/market/%s", marketId))
	}
	return "/"
}
