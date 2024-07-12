package handler

import (
	"git.ekzyis.com/ekzyis/delphi.market/server/router/context"
	"git.ekzyis.com/ekzyis/delphi.market/server/router/pages"
	"github.com/labstack/echo/v4"
)

func HandleIndex(sc context.Context) echo.HandlerFunc {
	return func(c echo.Context) error {
		return pages.Index().Render(context.RenderContext(sc, c), c.Response().Writer)
	}
}
