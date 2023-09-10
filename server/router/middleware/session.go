package middleware

import (
	"database/sql"
	"net/http"
	"time"

	"git.ekzyis.com/ekzyis/delphi.market/db"
	"github.com/labstack/echo/v4"
)

func Session(envVars map[string]any) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var (
				cookie *http.Cookie
				err    error
				s      *db.Session
				u      *db.User
			)
			if cookie, err = c.Cookie("session"); err != nil {
				// cookie not found
				return next(c)
			}
			s = &db.Session{SessionId: cookie.Value}
			if err = db.FetchSession(s); err == nil {
				// session found
				u = &db.User{Pubkey: s.Pubkey, LastSeen: time.Now()}
				if err = db.UpdateUser(u); err != nil {
					return err
				}
				c.Set("session", *u)
			} else if err != sql.ErrNoRows {
				return err
			}
			return next(c)
		}
	}
}

func SessionGuard(envVars map[string]any) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			session := c.Get("session")
			if session == nil {
				return c.Redirect(http.StatusTemporaryRedirect, "/login")
			}
			return next(c)
		}
	}
}
