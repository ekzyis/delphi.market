package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type User struct {
	Session
}

func user(c echo.Context) error {
	session := c.Get("session").(Session)
	u := User{}
	if err := db.FetchUser(session.Pubkey, &u); err != nil {
		return err
	}
	var orders []Order
	if err := db.FetchOrders(&FetchOrdersWhere{Pubkey: session.Pubkey}, &orders); err != nil {
		return err
	}
	data := map[string]any{
		"session": c.Get("session"),
		"user":    u,
		"Orders":  orders,
	}
	return c.Render(http.StatusOK, "user.html", data)
}
