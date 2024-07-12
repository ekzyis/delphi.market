package middleware

import (
	"net/http"

	"git.ekzyis.com/ekzyis/delphi.market/server/router/context"
	"github.com/labstack/echo/v4"
)

func Session(sc context.Context) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// TODO: implement session middleware
			// var (
			//   cookie *http.Cookie
			// 	 err    error
			// 	 s      *db.Session
			// 	 u      *db.User
			// )
			// if cookie, err = c.Cookie("session"); err != nil {
			// 	 // cookie not found
			// 	 return next(c)
			// }
			// s = &db.Session{SessionId: cookie.Value}
			// if err = sc.Db.FetchSession(s); err == nil {
			// 	 // session found
			// 	 c.Set("session", *u)
			// } else if err != sql.ErrNoRows {
			// 	 return err
			// }
			return next(c)
		}
	}
}

func SessionGuard(sc context.Context) echo.MiddlewareFunc {
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
