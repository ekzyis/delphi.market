package test

import (
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/labstack/echo/v4"
)

func HTTPMocks(method string, target string, body io.Reader) (*echo.Echo, *http.Request, *httptest.ResponseRecorder) {
	var (
		e   *echo.Echo
		req *http.Request
		rec *httptest.ResponseRecorder
	)
	e = echo.New()

	req = httptest.NewRequest(method, target, body)
	rec = httptest.NewRecorder()

	return e, req, rec
}
