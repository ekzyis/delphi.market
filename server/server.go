package server

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"git.ekzyis.com/ekzyis/delphi.market/server/router"
)

type Server struct {
	*echo.Echo
}

type ServerContext = router.ServerContext

func New(ctx ServerContext) *Server {
	var (
		e *echo.Echo
		s *Server
	)
	e = echo.New()
	e.Static("/", "public")
	e.Renderer = router.ParseTemplates("pages/**.html")
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format:           "${time_custom} ${method} ${uri} ${status}\n",
		CustomTimeFormat: "2006-01-02 15:04:05.00000-0700",
	}))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:4224", "https://delphi.market", "https://dev1.delphi.market"},
		AllowCredentials: true,
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	e.HTTPErrorHandler = httpErrorHandler

	s = &Server{e}

	router.AddRoutes(e, ctx)

	return s
}
