package handler

import (
	"database/sql"
	"net/http"
	"time"

	"git.ekzyis.com/ekzyis/delphi.market/db"
	"git.ekzyis.com/ekzyis/delphi.market/lib"
	"git.ekzyis.com/ekzyis/delphi.market/server/auth"
	"git.ekzyis.com/ekzyis/delphi.market/server/router/context"
	"github.com/labstack/echo/v4"
)

func HandleLogin(sc context.ServerContext) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			lnAuth   *auth.LNAuth
			dbLnAuth db.LNAuth
			err      error
			expires  time.Time = time.Now().Add(60 * 60 * 24 * 365 * time.Second)
			qr       string
			data     map[string]any
		)
		if lnAuth, err = auth.NewLNAuth(); err != nil {
			return err
		}
		dbLnAuth = db.LNAuth{K1: lnAuth.K1, LNURL: lnAuth.LNURL}
		if err = sc.Db.CreateLNAuth(&dbLnAuth); err != nil {
			return err
		}
		c.SetCookie(&http.Cookie{Name: "session", HttpOnly: true, Path: "/", Value: dbLnAuth.SessionId, Secure: true, Expires: expires})
		if qr, err = lib.ToQR(lnAuth.LNURL); err != nil {
			return err
		}
		data = map[string]any{
			"session": c.Get("session"),
			"lnurl":   lnAuth.LNURL,
			"qr":      qr,
		}
		return sc.Render(c, http.StatusOK, "login.html", data)
	}
}

func HandleLoginCallback(sc context.ServerContext) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			query     auth.LNAuthResponse
			sessionId string
			err       error
		)
		if err := c.Bind(&query); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest)
		}
		if err = sc.Db.FetchSessionId(query.K1, &sessionId); err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"reason": "session not found"})
		} else if err != nil {
			return err
		}
		if ok, err := auth.VerifyLNAuth(&query); err != nil {
			return err
		} else if !ok {
			return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"reason": "bad signature"})
		}
		if err = sc.Db.CreateUser(&db.User{Pubkey: query.Key}); err != nil {
			return err
		}
		if err = sc.Db.CreateSession(&db.Session{Pubkey: query.Key, SessionId: sessionId}); err != nil {
			return err
		}
		if err = sc.Db.DeleteLNAuth(&db.LNAuth{K1: query.K1}); err != nil {
			return err
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "OK"})
	}
}
