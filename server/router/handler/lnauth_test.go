package handler_test

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"git.ekzyis.com/ekzyis/delphi.market/server/auth"
	"git.ekzyis.com/ekzyis/delphi.market/server/router/context"
	"git.ekzyis.com/ekzyis/delphi.market/server/router/handler"
	"git.ekzyis.com/ekzyis/delphi.market/test"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func init() {
	test.Init(&db)
}

func TestLnAuthSignup(t *testing.T) {
	var (
		assert      = assert.New(t)
		sc          = context.Context{Db: db}
		e           *echo.Echo
		c           echo.Context
		req         *http.Request
		rec         *httptest.ResponseRecorder
		cookies     []*http.Cookie
		sessionId   string
		dbSessionId string
		err         error
	)

	e, req, rec = test.HTTPMocks("GET", "/signup/lightning", nil)
	c = e.NewContext(req, rec)
	c.SetParamNames("method")
	c.SetParamValues("lightning")

	err = handler.HandleAuth(sc, "register")(c)
	assert.NoErrorf(err, "handler returned error")

	// Set-Cookie header present
	cookies = rec.Result().Cookies()
	assert.Equalf(1, len(cookies), "wrong number of Set-Cookie headers")
	assert.Equalf("session", cookies[0].Name, "wrong cookie name")

	// new challenge inserted which matches cookie value
	sessionId = cookies[0].Value
	err = db.QueryRow("SELECT session_id FROM lnauth WHERE session_id = $1", sessionId).Scan(&dbSessionId)
	assert.NoError(err)
	assert.Equalf(sessionId, dbSessionId, "wrong session id")
}

func TestLnAuthSignupCallbackUserNotExists(t *testing.T) {
	var (
		assert    = assert.New(t)
		e         *echo.Echo
		c         echo.Context
		sc        context.Context
		req       *http.Request
		rec       *httptest.ResponseRecorder
		lnAuth    *auth.LnAuth
		sk        *secp256k1.PrivateKey
		pk        *secp256k1.PublicKey
		sig       string
		key       string
		sessionId string
		userId    int
		count     int
		err       error
	)

	lnAuth, err = auth.NewLnAuth("register")
	assert.NoErrorf(err, "error creating challenge")

	err = db.QueryRow(
		"INSERT INTO lnauth(k1, lnurl) VALUES($1, $2) RETURNING session_id",
		lnAuth.K1, lnAuth.LNURL).Scan(&sessionId)
	assert.NoErrorf(err, "error creating challenge")

	sk, pk, err = test.GenerateKeyPair()
	assert.NoErrorf(err, "error generating keypair")

	sig, err = test.Sign(sk, lnAuth.K1)
	assert.NoErrorf(err, "error signing k1")

	key = hex.EncodeToString(pk.SerializeCompressed())

	sc = context.Context{Db: db}
	e, req, rec = test.HTTPMocks("GET",
		fmt.Sprintf("/api/login?tag=login&k1=%s&key=%s&sig=%s&action=%s", lnAuth.K1, key, sig, "register"), nil)
	c = e.NewContext(req, rec)

	err = handler.HandleLnAuthCallback(sc)(c)
	assert.NoErrorf(err, "handler returned error")

	// user created
	err = db.QueryRow("SELECT id FROM users WHERE ln_pubkey = $1", key).Scan(&userId)
	assert.NoErrorf(err, "error fetching user")

	// session created
	err = db.QueryRow("SELECT COUNT(1) FROM sessions WHERE id = $1 AND user_id = $2", sessionId, userId).Scan(&count)
	assert.NoErrorf(err, "error fetching session")
	assert.Equalf(1, count, "invalid session count")

	// challenge deleted
	err = db.QueryRow("SELECT COUNT(1) FROM lnauth WHERE k1 = $1", lnAuth.K1).Scan(&count)
	assert.NoErrorf(err, "error fetching challenge")
	assert.Equalf(count, 0, "challenge not deleted")
}

func TestLnAuthSignupCallbackUserExists(t *testing.T) {
	var (
		assert    = assert.New(t)
		e         *echo.Echo
		c         echo.Context
		sc        context.Context
		req       *http.Request
		rec       *httptest.ResponseRecorder
		lnAuth    *auth.LnAuth
		sk        *secp256k1.PrivateKey
		pk        *secp256k1.PublicKey
		sig       string
		key       string
		sessionId string
		err       error
	)

	lnAuth, err = auth.NewLnAuth("register")
	assert.NoErrorf(err, "error creating challenge")

	err = db.QueryRow(
		"INSERT INTO lnauth(k1, lnurl) VALUES($1, $2) RETURNING session_id",
		lnAuth.K1, lnAuth.LNURL).Scan(&sessionId)
	assert.NoErrorf(err, "error creating challenge")

	sk, pk, err = test.GenerateKeyPair()
	assert.NoErrorf(err, "error generating keypair")

	sig, err = test.Sign(sk, lnAuth.K1)
	assert.NoErrorf(err, "error signing k1")

	key = hex.EncodeToString(pk.SerializeCompressed())

	// create user such that signup must fail
	_, err = db.Exec("INSERT INTO users(ln_pubkey) VALUES($1) RETURNING id", key)
	assert.NoErrorf(err, "error creating user")

	sc = context.Context{Db: db}
	e, req, rec = test.HTTPMocks("GET",
		fmt.Sprintf("/api/login?tag=login&k1=%s&key=%s&sig=%s&action=%s", lnAuth.K1, key, sig, "register"), nil)
	c = e.NewContext(req, rec)

	// must throw error because user already exists
	err = handler.HandleLnAuthCallback(sc)(c)
	assert.ErrorContains(err, "user already exists", "user check failed")
}

func TestLnAuthLogin(t *testing.T) {
	var (
		assert      = assert.New(t)
		sc          = context.Context{Db: db}
		e           *echo.Echo
		c           echo.Context
		req         *http.Request
		rec         *httptest.ResponseRecorder
		cookies     []*http.Cookie
		sessionId   string
		dbSessionId string
		err         error
	)

	e, req, rec = test.HTTPMocks("GET", "/login/lightning", nil)
	c = e.NewContext(req, rec)
	c.SetParamNames("method")
	c.SetParamValues("lightning")

	err = handler.HandleAuth(sc, "login")(c)
	assert.NoErrorf(err, "handler returned error")

	// Set-Cookie header present
	cookies = rec.Result().Cookies()
	assert.Equalf(len(cookies), 1, "wrong number of Set-Cookie headers")
	assert.Equalf("session", cookies[0].Name, "wrong cookie name")

	// new challenge inserted which matches cookie value
	sessionId = cookies[0].Value
	err = db.QueryRow("SELECT session_id FROM lnauth WHERE session_id = $1", sessionId).Scan(&dbSessionId)
	assert.NoError(err)
	assert.Equalf(sessionId, dbSessionId, "wrong session id")
}

func TestLnAuthLoginCallbackUserNotExists(t *testing.T) {
	var (
		assert    = assert.New(t)
		e         *echo.Echo
		c         echo.Context
		sc        context.Context
		req       *http.Request
		rec       *httptest.ResponseRecorder
		lnAuth    *auth.LnAuth
		sk        *secp256k1.PrivateKey
		pk        *secp256k1.PublicKey
		sig       string
		key       string
		sessionId string
		err       error
	)

	lnAuth, err = auth.NewLnAuth("login")
	assert.NoErrorf(err, "error creating challenge")

	err = db.QueryRow(
		"INSERT INTO lnauth(k1, lnurl) VALUES($1, $2) RETURNING session_id",
		lnAuth.K1, lnAuth.LNURL).Scan(&sessionId)
	assert.NoErrorf(err, "error creating challenge")

	sk, pk, err = test.GenerateKeyPair()
	assert.NoErrorf(err, "error generating keypair")

	sig, err = test.Sign(sk, lnAuth.K1)
	assert.NoErrorf(err, "error signing k1")

	key = hex.EncodeToString(pk.SerializeCompressed())

	sc = context.Context{Db: db}
	e, req, rec = test.HTTPMocks("GET",
		fmt.Sprintf("/api/login?tag=login&k1=%s&key=%s&sig=%s&action=%s", lnAuth.K1, key, sig, "login"), nil)
	c = e.NewContext(req, rec)

	// must throw error because user does not exist
	err = handler.HandleLnAuthCallback(sc)(c)
	assert.ErrorContains(err, "user not found", "user check failed")
}

func TestLnAuthLoginCallbackUserExists(t *testing.T) {
	var (
		assert    = assert.New(t)
		e         *echo.Echo
		c         echo.Context
		sc        context.Context
		req       *http.Request
		rec       *httptest.ResponseRecorder
		lnAuth    *auth.LnAuth
		sk        *secp256k1.PrivateKey
		pk        *secp256k1.PublicKey
		sig       string
		key       string
		sessionId string
		userId    int
		count     int
		err       error
	)

	lnAuth, err = auth.NewLnAuth("login")
	assert.NoErrorf(err, "error creating challenge")

	err = db.QueryRow(
		"INSERT INTO lnauth(k1, lnurl) VALUES($1, $2) RETURNING session_id",
		lnAuth.K1, lnAuth.LNURL).Scan(&sessionId)
	assert.NoErrorf(err, "error creating challenge")

	sk, pk, err = test.GenerateKeyPair()
	assert.NoErrorf(err, "error generating keypair")

	sig, err = test.Sign(sk, lnAuth.K1)
	assert.NoErrorf(err, "error signing k1")

	key = hex.EncodeToString(pk.SerializeCompressed())

	// create user such that login does not fail
	err = db.QueryRow("INSERT INTO users(ln_pubkey) VALUES($1) RETURNING id", key).Scan(&userId)
	assert.NoErrorf(err, "error creating user")

	sc = context.Context{Db: db}
	e, req, rec = test.HTTPMocks("GET",
		fmt.Sprintf("/api/login?tag=login&k1=%s&key=%s&sig=%s&action=%s", lnAuth.K1, key, sig, "login"), nil)
	c = e.NewContext(req, rec)

	err = handler.HandleLnAuthCallback(sc)(c)
	assert.NoErrorf(err, "handler returned error")

	// session created
	err = db.QueryRow("SELECT COUNT(1) FROM sessions WHERE id = $1 AND user_id = $2", sessionId, userId).Scan(&count)
	assert.NoErrorf(err, "error fetching session")
	assert.Equalf(1, count, "invalid session count")

	// challenge deleted
	err = db.QueryRow("SELECT COUNT(1) FROM lnauth WHERE k1 = $1", lnAuth.K1).Scan(&count)
	assert.NoErrorf(err, "error fetching challenge")
	assert.Equalf(count, 0, "challenge not deleted")
}
