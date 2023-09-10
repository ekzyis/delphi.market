package handler

import (
	"database/sql"
	"net/http"

	"git.ekzyis.com/ekzyis/delphi.market/db"
	"github.com/labstack/echo/v4"
)

func HandleCheckSession(envVars map[string]any) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			cookie *http.Cookie
			s      db.Session
			err    error
		)
		if cookie, err = c.Cookie("session"); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"reason": "cookie required"})
		}
		s = db.Session{SessionId: cookie.Value}
		if err = db.FetchSession(&s); err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"reason": "session not found"})
		} else if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		return c.JSON(http.StatusOK, map[string]string{"pubkey": s.Pubkey})
	}
}
