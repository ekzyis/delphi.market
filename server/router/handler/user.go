package handler

import (
	"net/http"

	"git.ekzyis.com/ekzyis/delphi.market/db"
	"git.ekzyis.com/ekzyis/delphi.market/server/router/context"
	"github.com/labstack/echo/v4"
)

func HandleUser(sc context.ServerContext) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			u      db.User
			orders []db.Order
			err    error
			data   map[string]any
		)
		u = c.Get("session").(db.User)
		if err = sc.Db.FetchOrders(&db.FetchOrdersWhere{Pubkey: u.Pubkey}, &orders); err != nil {
			return err
		}
		data = map[string]any{
			"session": c.Get("session"),
			"user":    u,
			"Orders":  orders,
		}
		return sc.Render(c, http.StatusOK, "user.html", data)
	}
}
