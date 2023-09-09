package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func index(c echo.Context) error {
	return c.Render(http.StatusOK, "index.html", map[string]string{"VERSION": VERSION, "COMMIT_LONG_SHA": COMMIT_LONG_SHA})
}

func login(c echo.Context) error {
	return c.Render(http.StatusOK, "login.html", map[string]string{"user": ""})
}

func serve500(c echo.Context) {
	f, err := os.Open("public/500.html")
	if err != nil {
		c.Logger().Error(err)
		return
	}
	err = c.Stream(500, "text/html", f)
	if err != nil {
		c.Logger().Error(err)
		return
	}
}

func httpErrorHandler(err error, c echo.Context) {
	c.Logger().Error(err)
	code := http.StatusInternalServerError
	httpError, ok := err.(*echo.HTTPError)
	if ok {
		code = httpError.Code
	}
	filePath := fmt.Sprintf("public/%d.html", code)
	f, err := os.Open(filePath)
	if err != nil {
		c.Logger().Error(err)
		serve500(c)
		return
	}
	err = c.Stream(code, "text/html", f)
	if err != nil {
		c.Logger().Error(err)
		serve500(c)
		return
	}
}
