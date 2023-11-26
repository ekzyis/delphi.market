package router

import (
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
	"golang.org/x/exp/constraints"
)

type Template struct {
	templates *template.Template
}

type Number interface {
	constraints.Integer | constraints.Float
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func ParseTemplates(pattern string) *Template {
	return &Template{
		templates: template.Must(template.New("").Funcs(template.FuncMap{
			"add":    add[int64],
			"sub":    sub[int64],
			"div":    div[int64],
			"substr": substr,
		}).ParseGlob("pages/**.html")),
	}
}

func add[T Number](arg1 T, arg2 T) T {
	return arg1 + arg2
}

func sub[T Number](arg1 T, arg2 T) T {
	return arg1 - arg2
}

func div[T Number](arg1 T, arg2 T) T {
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
