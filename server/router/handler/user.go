package handler

import (
	"git.ekzyis.com/ekzyis/delphi.market/server/router/context"
	"git.ekzyis.com/ekzyis/delphi.market/server/router/pages"
	"git.ekzyis.com/ekzyis/delphi.market/types"
	"github.com/labstack/echo/v4"
)

func HandleUser(sc context.Context) echo.HandlerFunc {
	return func(c echo.Context) error {
		u := c.Get("session").(types.User)
		return pages.User(&u).Render(context.RenderContext(sc, c), c.Response().Writer)
	}
}
