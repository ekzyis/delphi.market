package handler

import (
	"database/sql"
	"net/http"
	"time"

	"git.ekzyis.com/ekzyis/delphi.market/lib"
	"git.ekzyis.com/ekzyis/delphi.market/server/auth"
	"git.ekzyis.com/ekzyis/delphi.market/server/router/context"
	"git.ekzyis.com/ekzyis/delphi.market/server/router/pages"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
)

func HandleAuth(sc context.Context, action string) echo.HandlerFunc {
	return func(c echo.Context) error {

		if c.Param("method") == "lightning" {
			return LnAuth(sc, c, action)
		}

		return pages.Auth(mapAction(action)).Render(context.RenderContext(sc, c), c.Response().Writer)
	}
}

func LnAuth(sc context.Context, c echo.Context, action string) error {
	var (
		db        = sc.Db
		ctx       = c.Request().Context()
		lnAuth    *auth.LnAuth
		sessionId string
		// sessions expire in 30 days. TODO: refresh sessions
		expires = time.Now().Add(60 * 60 * 24 * 30 * time.Second)
		qr      string
		err     error
	)

	if lnAuth, err = auth.NewLnAuth(action); err != nil {
		return err
	}

	if err = db.QueryRowContext(
		ctx,
		"INSERT INTO lnauth(k1, lnurl) VALUES($1, $2) RETURNING session_id",
		lnAuth.K1, lnAuth.LNURL).Scan(&sessionId); err != nil {
		return err
	}

	if qr, err = lib.ToQR(lnAuth.LNURL); err != nil {
		return err
	}

	c.SetCookie(&http.Cookie{
		Name:     "session",
		HttpOnly: true,
		Path:     "/",
		Value:    sessionId,
		Secure:   true,
		Expires:  expires,
	})

	return pages.LnAuth(qr, lnAuth.LNURL, mapAction(action)).Render(context.RenderContext(sc, c), c.Response().Writer)
}

func HandleLnAuthCallback(sc context.Context) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			db        = sc.Db
			tx        *sql.Tx
			ctx       = c.Request().Context()
			query     auth.LnAuthCallback
			sessionId string
			userId    int
			ok        bool
			err       error
			pqErr     *pq.Error
		)

		bail := func(code int, reason string) error {
			c.JSON(code, map[string]string{"status": "ERROR", "reason": reason})
			return nil
		}

		if err = c.Bind(&query); err != nil {
			return bail(http.StatusInternalServerError, err.Error())
		} else if query.K1 == "" || query.Sig == "" || query.Key == "" || query.Action == "" {
			return bail(http.StatusBadRequest, "bad query")
		}

		if tx, err = db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted}); err != nil {
			return bail(http.StatusInternalServerError, err.Error())
		}

		err = tx.QueryRow("SELECT session_id FROM lnauth WHERE k1 = $1 LIMIT 1", query.K1).Scan(&sessionId)
		if err == sql.ErrNoRows {
			tx.Rollback()
			return bail(http.StatusNotFound, "session not found")
		} else if err != nil {
			tx.Rollback()
			return bail(http.StatusInternalServerError, err.Error())
		}

		ok, err = auth.VerifyLNAuth(&query)
		if err != nil {
			tx.Rollback()
			return bail(http.StatusInternalServerError, err.Error())
		} else if !ok {
			tx.Rollback()
			return bail(http.StatusBadRequest, "bad signature")
		}

		if query.Action == "register" {
			err = tx.QueryRow("INSERT INTO users(ln_pubkey) VALUES ($1) RETURNING id", query.Key).Scan(&userId)
			if err != nil {
				tx.Rollback()
				pqErr, ok = err.(*pq.Error)
				if ok && pqErr.Code == "23505" {
					return bail(http.StatusBadRequest, "user already exists")
				}
				return bail(http.StatusInternalServerError, err.Error())
			}
		} else if query.Action == "login" {
			err = tx.QueryRow("SELECT id FROM users WHERE ln_pubkey = $1", query.Key).Scan(&userId)
			if err == sql.ErrNoRows {
				tx.Rollback()
				return bail(http.StatusNotFound, "user not found")
			} else if err != nil {
				tx.Rollback()
				return bail(http.StatusInternalServerError, err.Error())
			}
		} else {
			return bail(http.StatusBadRequest, "bad action")
		}

		if _, err = tx.Exec("INSERT INTO sessions(id, user_id) VALUES($1, $2)", sessionId, userId); err != nil {
			tx.Rollback()
			return bail(http.StatusInternalServerError, err.Error())
		}

		if _, err = tx.Exec("DELETE FROM lnauth WHERE k1 = $1", query.K1); err != nil {
			tx.Rollback()
			return bail(http.StatusInternalServerError, err.Error())
		}

		if err = tx.Commit(); err != nil {
			tx.Rollback()
			return bail(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, map[string]string{"status": "OK"})
	}
}

func HandleSessionCheck(sc context.Context) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			db     = sc.Db
			ctx    = c.Request().Context()
			cookie *http.Cookie
			userId int
			err    error
		)

		if cookie, err = c.Cookie("session"); err != nil {
			return c.JSON(http.StatusUnauthorized, "no session cookie")
		}

		if err = db.QueryRowContext(ctx,
			"SELECT user_id FROM sessions WHERE id = $1", cookie.Value).
			Scan(&userId); err != nil {
			return c.JSON(http.StatusNotFound, "session not found")
		}

		c.Response().Header().Set("HX-Location", "/")
		// htmx requires a 200 response to follow redirects
		// see https://github.com/bigskysoftware/htmx/issues/2052
		return c.HTML(http.StatusOK, "/")
	}
}

func mapAction(action string) string {
	// LNURL spec uses "register" but we want to show "signup" to the user
	// see https://github.com/lnurl/luds/blob/luds/04.md
	switch action {
	case "register":
		return "signup"
	default:
		return action
	}
}
