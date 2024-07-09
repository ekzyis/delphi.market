package handler

import (
	"net/http"

	"git.ekzyis.com/ekzyis/delphi.market/db"
	"git.ekzyis.com/ekzyis/delphi.market/server/router/context"
	"git.ekzyis.com/ekzyis/delphi.market/server/router/pages"
	"github.com/labstack/echo/v4"
)

func HandleIndex(sc context.Context) echo.HandlerFunc {
	return func(c echo.Context) error {
		return pages.Index().Render(context.RenderContext(sc, c), c.Response().Writer)
	}
}

func HandleMarkets(sc context.Context) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			markets []db.Market
			err     error
		)
		if err = sc.Db.FetchActiveMarkets(&markets); err != nil {
			return err
		}
		return c.JSON(http.StatusOK, markets)
	}
}
