package handler

import (
	"net/http"

	"git.ekzyis.com/ekzyis/delphi.market/db"
	"git.ekzyis.com/ekzyis/delphi.market/lib"
	"github.com/labstack/echo/v4"
)

func HandleIndex(envVars map[string]any) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			markets []db.Market
			err     error
			data    map[string]any
		)
		if err = db.FetchActiveMarkets(&markets); err != nil {
			return err
		}
		data = map[string]any{
			"session": c.Get("session"),
			"markets": markets,
		}
		lib.Merge(&data, &envVars)
		return c.Render(http.StatusOK, "index.html", data)
	}
}
