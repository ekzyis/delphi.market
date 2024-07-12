package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"git.ekzyis.com/ekzyis/delphi.market/server/router/context"
	"git.ekzyis.com/ekzyis/delphi.market/server/router/handler"
	"git.ekzyis.com/ekzyis/delphi.market/test"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func init() {
	test.Init(&db)
}

func TestLnAuth(t *testing.T) {
	var (
		assert      = assert.New(t)
		e           *echo.Echo
		c           echo.Context
		sc          context.Context
		req         *http.Request
		rec         *httptest.ResponseRecorder
		cookies     []*http.Cookie
		sessionId   string
		dbSessionId string
		err         error
	)
	sc = context.Context{Db: db}
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

func TestLnAuthCallback(t *testing.T) {
	t.Skip()
}
