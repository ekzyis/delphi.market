package handler

import (
	"fmt"
	"net/http"
	"regexp"

	"git.ekzyis.com/ekzyis/delphi.market/server/router/context"
	"git.ekzyis.com/ekzyis/delphi.market/server/router/pages"
	"git.ekzyis.com/ekzyis/delphi.market/types"
	"github.com/labstack/echo/v4"
)

func HandleUser(sc context.Context) echo.HandlerFunc {
	return func(c echo.Context) error {
		u := c.Get("session").(types.User)
		return pages.User(&u, nil).Render(context.RenderContext(sc, c), c.Response().Writer)
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
			errors    types.UserEditError
			err       error
		)

		if name == "" {
			errors.Name = "required"
		} else if len(name) > maxLength {
			errors.Name = fmt.Sprintf("%d characters too long", len(name)-maxLength)
		} else if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(name) {
			errors.Name = "only letters, numbers, _ and - allowed"
		}

		if errors.Name != "" {
			c.Response().WriteHeader(http.StatusBadRequest)
			return pages.User(&u, &errors).Render(context.RenderContext(sc, c), c.Response().Writer)
		}

		if err = db.QueryRowContext(ctx,
			"UPDATE users SET name = $1 WHERE id = $2 RETURNING name",
			name, u.Id).Scan(&u.Name); err != nil {
			return err
		}

		return pages.User(&u, &errors).Render(context.RenderContext(sc, c), c.Response().Writer)
	}
}
