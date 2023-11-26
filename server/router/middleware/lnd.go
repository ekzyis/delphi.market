package middleware

import (
	"net/http"

	"git.ekzyis.com/ekzyis/delphi.market/server/router/context"
	"github.com/labstack/echo/v4"
)

func LNDGuard(sc context.ServerContext) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if sc.Lnd != nil {
				return next(c)
			}
			return echo.NewHTTPError(http.StatusMethodNotAllowed)
		}
	}
}
