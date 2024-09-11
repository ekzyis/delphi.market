package handler

import (
	"fmt"
	"net/http"

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

func HandleUserEdit(sc context.Context) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			db   = sc.Db
			ctx  = c.Request().Context()
			u    = c.Get("session").(types.User)
			name = c.FormValue("name")

			maxLength = 16
			err       error
		)

		if name == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "name is required")
		}

		if len(name) > maxLength {
			echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("name cannot be longer than %d characters", maxLength))
		}

		if err = db.QueryRowContext(ctx,
			"UPDATE users SET name = $1 WHERE id = $2 RETURNING name",
			name, u.Id).Scan(&u.Name); err != nil {
			return err
		}

		return pages.User(&u).Render(context.RenderContext(sc, c), c.Response().Writer)
	}
}
