package handler

import (
	"database/sql"
	"net/http"

	"git.ekzyis.com/ekzyis/delphi.market/db"
	"git.ekzyis.com/ekzyis/delphi.market/server/router/context"
	"github.com/labstack/echo/v4"
)

func HandleCheckSession(sc context.ServerContext) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			cookie *http.Cookie
			s      db.Session
			err    error
		)
		if cookie, err = c.Cookie("session"); err != nil {
			// return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"reason": "cookie required"})
			return c.JSON(http.StatusBadRequest, map[string]string{"reason": "cookie required"})
		}
		s = db.Session{SessionId: cookie.Value}
		if err = sc.Db.FetchSession(&s); err == sql.ErrNoRows {
			// return echo.NewHTTPError(http.StatusNotFound, map[string]string{"reason": "session not found"})
			return c.JSON(http.StatusBadRequest, map[string]string{"reason": "session not found"})
		} else if err != nil {
			// return echo.NewHTTPError(http.StatusInternalServerError)
			return c.JSON(http.StatusInternalServerError, nil)
		}
		return c.JSON(http.StatusOK, map[string]string{"pubkey": s.Pubkey})
	}
}
