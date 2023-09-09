package main

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/skip2/go-qrcode"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func index(c echo.Context) error {
	return c.Render(http.StatusOK, "index.html", map[string]any{"session": c.Get("session"), "VERSION": VERSION, "COMMIT_LONG_SHA": COMMIT_LONG_SHA})
}

func login(c echo.Context) error {
	lnauth, err := lnAuth()
	if err != nil {
		return err
	}
	var sessionId string
	err = db.QueryRow("INSERT INTO lnauth(k1, lnurl) VALUES($1, $2) RETURNING session_id", lnauth.k1, lnauth.lnurl).Scan(&sessionId)
	if err != nil {
		return err
	}
	expires := time.Now().Add(60 * 60 * 24 * 365 * time.Second)
	c.SetCookie(&http.Cookie{Name: "session", HttpOnly: true, Path: "/", Value: sessionId, Secure: true, Expires: expires})
	png, err := qrcode.Encode(lnauth.lnurl, qrcode.Medium, 256)
	if err != nil {
		return err
	}
	qr := base64.StdEncoding.EncodeToString([]byte(png))
	return c.Render(http.StatusOK, "login.html", map[string]any{"session": c.Get("session"), "lnurl": lnauth.lnurl, "qr": qr})
}

func verifyLogin(c echo.Context) error {
	var query LnAuthResponse
	if err := c.Bind(&query); err != nil {
		c.Logger().Error(err)
		return c.JSON(http.StatusBadRequest, map[string]string{"status": "ERROR", "reason": "bad request"})
	}
	var sessionId string
	err := db.QueryRow("SELECT session_id FROM lnauth WHERE k1 = $1", query.K1).Scan(&sessionId)
	if err == sql.ErrNoRows {
		return c.JSON(http.StatusBadRequest, map[string]string{"status": "ERROR", "reason": "unknown k1"})
	} else if err != nil {
		c.Logger().Error(err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"status": "ERROR", "reason": "internal server error"})
	}
	ok, err := lnAuthVerify(&query)
	if err != nil {
		c.Logger().Error(err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"status": "ERROR", "reason": "internal server error"})
	}
	if !ok {
		c.Logger().Error("bad signature")
		return c.JSON(http.StatusUnauthorized, map[string]string{"status": "ERROR", "reason": "bad signature"})
	}
	_, err = db.Exec("INSERT INTO sessions(pubkey, session_id) VALUES($1, $2)", query.Key, sessionId)
	if err != nil {
		c.Logger().Error(err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"status": "ERROR", "reason": "internal server error"})
	}
	_, err = db.Exec("DELETE FROM lnauth WHERE k1 = $1", query.K1)
	if err != nil {
		c.Logger().Error(err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"status": "ERROR", "reason": "internal server error"})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "OK"})
}

func checkSession(c echo.Context) error {
	cookie, err := c.Cookie("session")
	if err != nil {
		c.Logger().Error(err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"status": "ERROR", "reason": "internal server error"})
	}
	sessionId := cookie.Value
	var pubkey string
	err = db.QueryRow("SELECT pubkey FROM sessions WHERE session_id = $1", sessionId).Scan(&pubkey)
	if err == sql.ErrNoRows {
		return c.JSON(http.StatusNotFound, map[string]string{"status": "Not Found", "message": "session not found"})
	} else if err != nil {
		c.Logger().Error(err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"status": "ERROR", "reason": "internal server error"})
	}
	return c.JSON(http.StatusOK, map[string]string{"pubkey": pubkey})
}

func sessionHandler(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie("session")
		if err != nil {
			// cookie not found
			return next(c)
		}
		sessionId := cookie.Value
		var pubkey string
		err = db.QueryRow("SELECT pubkey FROM sessions WHERE session_id = $1", sessionId).Scan(&pubkey)
		if err == nil {
			// session found
			c.Set("session", Session{pubkey})
		} else if err != sql.ErrNoRows {
			c.Logger().Error(err)
		}
		return next(c)
	}
}

func logout(c echo.Context) error {
	cookie, err := c.Cookie("session")
	if err != nil {
		// cookie not found
		return c.Redirect(http.StatusSeeOther, "/")
	}
	sessionId := cookie.Value
	_, err = db.Exec("DELETE FROM sessions where session_id = $1", sessionId)
	if err != nil {
		c.Logger().Error(err)
		return err
	}
	// tell browser that cookie is expired and thus can be deleted
	c.SetCookie(&http.Cookie{Name: "session", HttpOnly: true, Path: "/", Value: sessionId, Secure: true, Expires: time.Now()})
	return c.Redirect(http.StatusSeeOther, "/")
}

func serve500(c echo.Context) {
	f, err := os.Open("public/500.html")
	if err != nil {
		c.Logger().Error(err)
		return
	}
	err = c.Stream(500, "text/html", f)
	if err != nil {
		c.Logger().Error(err)
		return
	}
}

func httpErrorHandler(err error, c echo.Context) {
	c.Logger().Error(err)
	code := http.StatusInternalServerError
	httpError, ok := err.(*echo.HTTPError)
	if ok {
		code = httpError.Code
	}
	filePath := fmt.Sprintf("public/%d.html", code)
	f, err := os.Open(filePath)
	if err != nil {
		c.Logger().Error(err)
		serve500(c)
		return
	}
	err = c.Stream(code, "text/html", f)
	if err != nil {
		c.Logger().Error(err)
		serve500(c)
		return
	}
}
