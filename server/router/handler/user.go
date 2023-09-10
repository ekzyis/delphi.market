package handler

import (
	"net/http"

	"git.ekzyis.com/ekzyis/delphi.market/db"
	"git.ekzyis.com/ekzyis/delphi.market/lib"
	"github.com/labstack/echo/v4"
)

func HandleUser(envVars map[string]any) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			u      db.User
			orders []db.Order
			err    error
			data   map[string]any
		)
		u = c.Get("session").(db.User)
		if err = db.FetchOrders(&db.FetchOrdersWhere{Pubkey: u.Pubkey}, &orders); err != nil {
			return err
		}
		data = map[string]any{
			"session": c.Get("session"),
			"user":    u,
			"Orders":  orders,
		}
		lib.Merge(&data, &envVars)
		return c.Render(http.StatusOK, "user.html", data)
	}
}
