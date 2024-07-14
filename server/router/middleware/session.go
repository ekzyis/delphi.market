package middleware

import (
	"database/sql"
	"net/http"

	"git.ekzyis.com/ekzyis/delphi.market/server/router/context"
	"git.ekzyis.com/ekzyis/delphi.market/types"
	"github.com/labstack/echo/v4"
)

func Session(sc context.Context) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var (
				db     = sc.Db
				ctx    = c.Request().Context()
				cookie *http.Cookie
				err    error
				u      = types.User{}
			)
			if cookie, err = c.Cookie("session"); err != nil {
				// cookie not found
				return next(c)
			}
			if err = db.QueryRowContext(
				ctx,
				""+
					"SELECT u.id, u.created_at, COALESCE(u.ln_pubkey, ''), COALESCE(u.nostr_pubkey, ''), u.msats "+
					"FROM sessions s LEFT JOIN users u ON u.id = s.user_id "+
					"WHERE s.id = $1",
				cookie.Value).
				Scan(&u.Id, &u.CreatedAt, &u.LnPubkey, &u.NostrPubkey, &u.Msats); err == nil {
				// session found
				c.Set("session", u)
			} else if err != sql.ErrNoRows {
				return err
			}
			return next(c)
		}
	}
}

func SessionGuard(sc context.Context) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			session := c.Get("session")
			if session == nil {
				// this seems to work for non-interactive and htmx requests
				return c.Redirect(http.StatusTemporaryRedirect, "/login")
			}
			return next(c)
		}
	}
}
