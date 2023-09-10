package handler

import (
	"net/http"
	"time"

	"git.ekzyis.com/ekzyis/delphi.market/db"
	"github.com/labstack/echo/v4"
)

func HandleLogout(envVars map[string]any) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			cookie    *http.Cookie
			sessionId string
			err       error
		)
		if cookie, err = c.Cookie("session"); err != nil {
			// cookie not found
			return c.Redirect(http.StatusSeeOther, "/")
		}
		sessionId = cookie.Value
		if err = db.DeleteSession(&db.Session{SessionId: sessionId}); err != nil {
			return err
		}
		// tell browser that cookie is expired and thus can be deleted
		c.SetCookie(&http.Cookie{Name: "session", HttpOnly: true, Path: "/", Value: sessionId, Secure: true, Expires: time.Now()})
		return c.Redirect(http.StatusSeeOther, "/")
	}
}
