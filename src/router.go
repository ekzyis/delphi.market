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

func add(arg1 int, arg2 int) int {
	return arg1 + arg2
}

func sub(arg1 int, arg2 int) int {
	return arg1 - arg2
}

func div(arg1 int, arg2 int) int {
	return arg1 / arg2
}

func substr(s string, start, length int) string {
	if start < 0 || start >= len(s) {
		return ""
	}
	end := start + length
	if end > len(s) {
		end = len(s)
	}
	return s[start:end]
}

var (
	FuncMap template.FuncMap = template.FuncMap{
		"add":    add,
		"sub":    sub,
		"div":    div,
		"substr": substr,
	}
)

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func index(c echo.Context) error {
	rows, err := db.Query("SELECT id, description, active FROM markets WHERE active = true")
	if err != nil {
		return err
	}
	defer rows.Close()
	var markets []Market
	for rows.Next() {
		var market Market
		rows.Scan(&market.Id, &market.Description, &market.Active)
		markets = append(markets, market)
	}
	data := map[string]any{
		"session":         c.Get("session"),
		"ENV":             ENV,
		"markets":         markets,
		"VERSION":         VERSION,
		"COMMIT_LONG_SHA": COMMIT_LONG_SHA}
	return c.Render(http.StatusOK, "index.html", data)
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
