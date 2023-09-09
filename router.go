package main

import (
	"html/template"
	"io"
	"net/http"

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
