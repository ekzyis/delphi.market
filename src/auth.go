package main

import (
	"crypto/ecdsa"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcutil/bech32"
	"github.com/labstack/echo/v4"
	"github.com/skip2/go-qrcode"
)

type LnAuth struct {
	k1    string
	lnurl string
}

type LnAuthResponse struct {
	K1  string `query:"k1"`
	Sig string `query:"sig"`
	Key string `query:"key"`
}

type Session struct {
	pubkey string
}

func lnAuth() (*LnAuth, error) {
	k1 := make([]byte, 32)
	_, err := rand.Read(k1)
	if err != nil {
		return nil, fmt.Errorf("rand.Read error: %w", err)
	}
	k1hex := hex.EncodeToString(k1)
	url := []byte(fmt.Sprintf("https://%s/api/login?tag=login&k1=%s&action=login", PUBLIC_URL, k1hex))
	conv, err := bech32.ConvertBits(url, 8, 5, true)
	if err != nil {
		return nil, fmt.Errorf("bech32.ConvertBits error: %w", err)
	}
	lnurl, err := bech32.Encode("lnurl", conv)
	if err != nil {
		return nil, fmt.Errorf("bech32.Encode error: %w", err)
	}
	return &LnAuth{k1hex, lnurl}, nil
}

func lnAuthVerify(r *LnAuthResponse) (bool, error) {
	var k1Bytes, sigBytes, keyBytes []byte
	k1Bytes, err := hex.DecodeString(r.K1)
	if err != nil {
		return false, fmt.Errorf("k1 decode error: %w", err)
	}
	sigBytes, err = hex.DecodeString(r.Sig)
	if err != nil {
		return false, fmt.Errorf("sig decode error: %w", err)
	}
	keyBytes, err = hex.DecodeString(r.Key)
	if err != nil {
		return false, fmt.Errorf("key decode error: %w", err)
	}
	key, err := btcec.ParsePubKey(keyBytes)
	if err != nil {
		return false, fmt.Errorf("key parse error: %w", err)
	}
	ecdsaKey := ecdsa.PublicKey{Curve: btcec.S256(), X: key.X(), Y: key.Y()}
	return ecdsa.VerifyASN1(&ecdsaKey, k1Bytes, sigBytes), nil
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
	return c.Render(http.StatusOK, "login.html", map[string]any{"session": c.Get("session"), "PUBLIC_URL": PUBLIC_URL, "lnurl": lnauth.lnurl, "qr": qr})
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
	_, err = db.Exec("INSERT INTO users(pubkey) VALUES ($1) ON CONFLICT(pubkey) DO UPDATE SET last_seen = CURRENT_TIMESTAMP", query.Key)
	if err != nil {
		c.Logger().Error(err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"status": "ERROR", "reason": "internal server error"})
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
			_, err = db.Exec("UPDATE users SET last_seen = CURRENT_TIMESTAMP WHERE pubkey = $1", pubkey)
			if err != nil {
				c.Logger().Error(err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"status": "ERROR", "reason": "internal server error"})
			}
			c.Set("session", Session{pubkey})
		} else if err != sql.ErrNoRows {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"status": "ERROR", "reason": "internal server error"})
		}
		return next(c)
	}
}

func sessionGuard(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session")
		if session == nil {
			return c.Redirect(http.StatusTemporaryRedirect, "/login")
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
