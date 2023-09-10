package server

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"git.ekzyis.com/ekzyis/delphi.market/server/router"
)

type Server struct {
	*echo.Echo
}

func NewServer() *Server {
	e := echo.New()

	e.Static("/", "public")

	e.Renderer = router.T

	router.AddRoutes(e)

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format:           "${time_custom} ${method} ${uri} ${status}\n",
		CustomTimeFormat: "2006-01-02 15:04:05.00000-0700",
	}))

	e.HTTPErrorHandler = httpErrorHandler

	return &Server{e}
}
