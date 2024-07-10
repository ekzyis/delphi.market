package router

import (
	"github.com/labstack/echo/v4"

	"git.ekzyis.com/ekzyis/delphi.market/server/router/context"
	"git.ekzyis.com/ekzyis/delphi.market/server/router/handler"
	"git.ekzyis.com/ekzyis/delphi.market/server/router/middleware"
)

type Context = context.Context

func Init(e *echo.Echo, sc Context) {
	e.Use(middleware.Session(sc))

	e.GET("/", handler.HandleIndex(sc))
	e.GET("/about", handler.HandleAbout(sc))
}
