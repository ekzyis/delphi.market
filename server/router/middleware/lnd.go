package middleware

import (
	"net/http"

	"git.ekzyis.com/ekzyis/delphi.market/lnd"
	"github.com/labstack/echo/v4"
)

func LNDGuard(envVars map[string]any) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if lnd.Enabled {
				return next(c)
			}
			return echo.NewHTTPError(http.StatusMethodNotAllowed)
		}
	}
}
