package server

import (
	"bytes"
	"net/http"
	"strings"

	"git.ekzyis.com/ekzyis/delphi.market/server/router/context"
	"git.ekzyis.com/ekzyis/delphi.market/server/router/pages"
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func httpErrorHandler(sc context.Context) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		var (
			code = http.StatusInternalServerError
			page = pages.Error
			buf  *bytes.Buffer
		)

		c.Logger().Error(err)

		if httpError, ok := err.(*echo.HTTPError); ok {
			code = httpError.Code
		}

		if strings.Contains(err.Error(), "violates check constraint") ||
			strings.Contains(err.Error(), "violates unique constraint") {
			code = 400
		}

		buf = templ.GetBuffer()
		defer templ.ReleaseBuffer(buf)

		if err = page(code).Render(context.RenderContext(sc, c), buf); err != nil {
			c.Logger().Error(err)
			code = http.StatusInternalServerError
		}

		if err = c.HTML(code, buf.String()); err != nil {
			c.Logger().Error(err)
		}
	}
}
