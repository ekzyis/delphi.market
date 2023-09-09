package main

import (
	"html/template"
)

func add(arg1 int, arg2 int) int {
	return arg1 + arg2
}

func sub(arg1 int, arg2 int) int {
	return arg1 - arg2
}

var (
	FuncMap template.FuncMap = template.FuncMap{
		"add": add,
		"sub": sub,
	}
)
