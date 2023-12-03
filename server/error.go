package server

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func httpErrorHandler(err error, c echo.Context) {
	c.Logger().Error(err)
	code := http.StatusInternalServerError
	if httpError, ok := err.(*echo.HTTPError); ok {
		code = httpError.Code
	}
	if strings.Contains(err.Error(), "violates check constraint") || strings.Contains(err.Error(), "violates unique constraint") {
		code = 400
	}
	c.JSON(code, map[string]any{"status": code})
}
