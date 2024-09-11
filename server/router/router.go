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
	e.GET("/create", handler.HandleIndex(sc))
	e.POST("/create", handler.HandleCreate(sc), middleware.SessionGuard(sc))
	e.GET("/market/:id", handler.HandleMarket(sc))
	e.POST("/market/:id/order", handler.HandleOrder(sc), middleware.SessionGuard(sc))
	e.GET("/about", handler.HandleAbout(sc))

	e.GET("/login", handler.HandleAuth(sc, "login"))
	e.GET("/login/:method", handler.HandleAuth(sc, "login"))
	e.GET("/signup", handler.HandleAuth(sc, "register"))
	e.GET("/signup/:method", handler.HandleAuth(sc, "register"))
	e.GET("/api/lnauth/callback", handler.HandleLnAuthCallback(sc))
	e.GET("/session", handler.HandleSessionCheck(sc))

	e.GET("/user", handler.HandleUser(sc), middleware.SessionGuard(sc))
	e.PUT("/user", handler.HandleUserEdit(sc), middleware.SessionGuard(sc))
	e.POST("/logout", handler.HandleLogout(sc), middleware.SessionGuard(sc))

	e.GET("/invoice/:hash", handler.HandleInvoice(sc), middleware.SessionGuard(sc))
}
