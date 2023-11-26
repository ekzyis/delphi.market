package handler_test

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	db_ "git.ekzyis.com/ekzyis/delphi.market/db"
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

func TestLogin(t *testing.T) {
	var (
		assert      = assert.New(t)
		e           *echo.Echo
		c           echo.Context
		sc          context.ServerContext
		req         *http.Request
		rec         *httptest.ResponseRecorder
		cookies     []*http.Cookie
		sessionId   string
		dbSessionId string
		err         error
	)
	sc = context.ServerContext{Db: db}
	e, req, rec = test.HTTPMocks("GET", "/login", nil)
	c = e.NewContext(req, rec)

	err = handler.HandleLogin(sc)(c)
	assert.NoErrorf(err, "handler returned error")

	// Set-Cookie header present
	cookies = rec.Result().Cookies()
	assert.Equalf(len(cookies), 1, "wrong number of Set-Cookie headers")
	assert.Equalf(cookies[0].Name, "session", "wrong cookie name")

	// new challenge inserted
	sessionId = cookies[0].Value
	err = db.QueryRow("SELECT session_id FROM lnauth WHERE session_id = $1", sessionId).Scan(&dbSessionId)
	if !assert.NoError(err) {
		return
	}

	// inserted challenge matches cookie value
	assert.Equalf(sessionId, dbSessionId, "wrong session id")
}

func TestLoginCallback(t *testing.T) {
	var (
		assert   = assert.New(t)
		e        *echo.Echo
		c        echo.Context
		sc       context.ServerContext
		req      *http.Request
		rec      *httptest.ResponseRecorder
		sk       *secp256k1.PrivateKey
		pk       *secp256k1.PublicKey
		lnAuth   *auth.LNAuth
		dbLnAuth *db_.LNAuth
		u        *db_.User
		s        *db_.Session
		key      string
		sig      string
		err      error
	)
	lnAuth, err = auth.NewLNAuth()
	if !assert.NoErrorf(err, "error creating challenge") {
		return
	}
	dbLnAuth = &db_.LNAuth{K1: lnAuth.K1, LNURL: lnAuth.LNURL}
	err = db.CreateLNAuth(dbLnAuth)
	if !assert.NoErrorf(err, "error inserting challenge") {
		return
	}

	sk, pk, err = test.GenerateKeyPair()
	if !assert.NoErrorf(err, "error generating keypair") {
		return
	}
	sig, err = test.Sign(sk, lnAuth.K1)
	if !assert.NoErrorf(err, "error signing k1") {
		return
	}
	key = hex.EncodeToString(pk.SerializeCompressed())

	sc = context.ServerContext{Db: db}
	e, req, rec = test.HTTPMocks("GET", fmt.Sprintf("/api/login?k1=%s&key=%s&sig=%s", lnAuth.K1, key, sig), nil)
	c = e.NewContext(req, rec)

	err = handler.HandleLoginCallback(sc)(c)
	assert.NoErrorf(err, "handler returned error")

	// user created
	u = new(db_.User)
	err = db.FetchUser(key, u)
	if assert.NoErrorf(err, "error fetching user") {
		assert.Equalf(u.Pubkey, key, "pubkeys do not match")
	}

	// session created
	s = &db_.Session{SessionId: dbLnAuth.SessionId}
	err = db.FetchSession(s)
	if assert.NoErrorf(err, "error fetching session") {
		assert.Equalf(s.Pubkey, u.Pubkey, "session pubkey does not match user pubkey")
	}

	// challenge deleted
	err = db.FetchSessionId(u.Pubkey, &dbLnAuth.SessionId)
	assert.ErrorIs(err, sql.ErrNoRows, "challenge not deleted")
}
