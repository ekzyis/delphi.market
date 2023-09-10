package handler

import (
	"database/sql"
	"net/http"
	"time"

	"git.ekzyis.com/ekzyis/delphi.market/db"
	"git.ekzyis.com/ekzyis/delphi.market/lib"
	"git.ekzyis.com/ekzyis/delphi.market/server/auth"
	"github.com/labstack/echo/v4"
)

func HandleLogin(envVars map[string]any) echo.HandlerFunc {
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
		if err = db.CreateLNAuth(&dbLnAuth); err != nil {
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
		lib.Merge(&data, &envVars)
		return c.Render(http.StatusOK, "login.html", data)
	}
}

func HandleLoginCallback(envVars map[string]any) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			query     auth.LNAuthResponse
			sessionId string
			err       error
		)
		if err := c.Bind(&query); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest)
		}
		if err = db.FetchSessionId(query.K1, &sessionId); err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"reason": "session not found"})
		} else if err != nil {
			return err
		}
		if ok, err := auth.VerifyLNAuth(&query); err != nil {
			return err
		} else if !ok {
			return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"reason": "bad signature"})
		}
		if err = db.CreateUser(&db.User{Pubkey: query.Key}); err != nil {
			return err
		}
		if err = db.CreateSession(&db.Session{Pubkey: query.Key, SessionId: sessionId}); err != nil {
			return err
		}
		if err = db.DeleteLNAuth(&db.LNAuth{K1: query.K1}); err != nil {
			return err
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "OK"})
	}
}
