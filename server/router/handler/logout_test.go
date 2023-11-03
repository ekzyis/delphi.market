package handler_test

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	db_ "git.ekzyis.com/ekzyis/delphi.market/db"
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

func TestLogout(t *testing.T) {
	var (
		assert = assert.New(t)
		e      *echo.Echo
		c      echo.Context
		sc     context.ServerContext
		req    *http.Request
		rec    *httptest.ResponseRecorder
		pk     *secp256k1.PublicKey
		s      *db_.Session
		key    string
		err    error
	)
	sc = context.ServerContext{Db: db}
	e, req, rec = test.HTTPMocks("POST", "/logout", nil)

	_, pk, err = test.GenerateKeyPair()
	if !assert.NoErrorf(err, "error generating keypair") {
		return
	}
	key = hex.EncodeToString(pk.SerializeCompressed())
	err = sc.Db.CreateUser(&db_.User{Pubkey: key})
	if !assert.NoErrorf(err, "error creating user") {
		return
	}
	s = &db_.Session{Pubkey: key}
	err = sc.Db.QueryRow("SELECT encode(gen_random_uuid()::text::bytea, 'base64')").Scan(&s.SessionId)
	if !assert.NoErrorf(err, "error creating session id") {
		return
	}

	// create session
	err = sc.Db.CreateSession(s)
	if !assert.NoErrorf(err, "error creating session") {
		return
	}

	// session authentication via cookie
	req.Header.Add("cookie", fmt.Sprintf("session=%s", s.SessionId))

	c = e.NewContext(req, rec)
	err = handler.HandleLogout(sc)(c)
	assert.NoErrorf(err, "handler returned error")

	// session must have been deleted
	err = sc.Db.FetchSession(s)
	assert.ErrorIsf(err, sql.ErrNoRows, "session not deleted")

}
