package handler_test

import (
	"crypto/ecdsa"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"runtime"
	"testing"

	db_ "git.ekzyis.com/ekzyis/delphi.market/db"
	"git.ekzyis.com/ekzyis/delphi.market/server/auth"
	"git.ekzyis.com/ekzyis/delphi.market/server/router"
	"git.ekzyis.com/ekzyis/delphi.market/server/router/context"
	"git.ekzyis.com/ekzyis/delphi.market/server/router/handler"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

var (
	dbName string = "delphi_test"
	dbUrl  string = fmt.Sprintf("postgres://delphi:delphi@localhost:5432/%s?sslmode=disable", dbName)
	db     *db_.DB
)

func init() {
	// for ParseTemplates to work, cwd needs to be project root
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../../..")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
	if db, err = db_.New(dbUrl); err != nil {
		panic(err)
	}
}

func TestMain(m *testing.M) {
	if err := db.Reset(dbName); err != nil {
		panic(err)
	}
	retCode := m.Run()
	if err := db.Clear(dbName); err != nil {
		panic(err)
	}
	os.Exit(retCode)
}

func mocks(method string, target string, body io.Reader) (*echo.Echo, context.ServerContext, *http.Request, *httptest.ResponseRecorder) {
	var (
		e   *echo.Echo
		sc  context.ServerContext
		req *http.Request
		rec *httptest.ResponseRecorder
	)
	e = echo.New()
	e.Renderer = router.ParseTemplates("pages/**.html")
	sc = context.ServerContext{
		Db: db,
	}
	req = httptest.NewRequest(method, target, body)
	rec = httptest.NewRecorder()
	return e, sc, req, rec
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
	e, sc, req, rec = mocks("GET", "/login", nil)
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

	key, sig, err = sign(lnAuth.K1)
	if !assert.NoErrorf(err, "error signing k1") {
		return
	}

	e, sc, req, rec = mocks("GET", fmt.Sprintf("/api/login?k1=%s&key=%s&sig=%s", lnAuth.K1, key, sig), nil)
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

func sign(k1_ string) (string, string, error) {
	var (
		sk  *secp256k1.PrivateKey
		k1  []byte
		sig []byte
		err error
	)
	if k1, err = hex.DecodeString(k1_); err != nil {
		return "", "", err
	}
	if sk, err = secp256k1.GeneratePrivateKey(); err != nil {
		return "", "", err
	}
	if sig, err = ecdsa.SignASN1(rand.Reader, sk.ToECDSA(), k1); err != nil {
		return "", "", err
	}
	return hex.EncodeToString(sk.PubKey().SerializeCompressed()), hex.EncodeToString(sig), nil
}
