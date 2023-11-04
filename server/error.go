package server

import (
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
)

func httpErrorHandler(err error, c echo.Context) {
	c.Logger().Error(err)
	code := http.StatusInternalServerError
	if httpError, ok := err.(*echo.HTTPError); ok {
		code = httpError.Code
	}
	filePath := fmt.Sprintf("public/%d.html", code)
	var f *os.File
	if f, err = os.Open(filePath); err != nil {
		c.Logger().Error(err)
		serveError(c, 500)
		return
	}
	if err = c.Stream(code, "text/html", f); err != nil {
		c.Logger().Error(err)
		serveError(c, 500)
		return
	}
}

func serveError(c echo.Context, code int) error {
	var (
		f   *os.File
		err error
	)
	if f, err = os.Open(fmt.Sprintf("public/%d.html", code)); err != nil {
		c.Logger().Error(err)
		return err
	}
	// TODO return errors in JSON
	if err = c.Stream(code, "text/html", f); err != nil {
		c.Logger().Error(err)
		return err
	}
	return nil
}
