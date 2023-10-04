package handler

import (
	"net/http"

	"git.ekzyis.com/ekzyis/delphi.market/db"
	"git.ekzyis.com/ekzyis/delphi.market/server/router/context"
	"github.com/labstack/echo/v4"
)

func HandleIndex(sc context.ServerContext) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			markets []db.Market
			err     error
			data    map[string]any
		)
		if err = sc.Db.FetchActiveMarkets(&markets); err != nil {
			return err
		}
		data = map[string]any{
			"session": c.Get("session"),
			"markets": markets,
		}

		return sc.Render(c, http.StatusOK, "index.html", data)
	}
}
