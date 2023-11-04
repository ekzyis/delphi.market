package router

import (
	"github.com/labstack/echo/v4"

	"git.ekzyis.com/ekzyis/delphi.market/server/router/context"
	"git.ekzyis.com/ekzyis/delphi.market/server/router/handler"
	"git.ekzyis.com/ekzyis/delphi.market/server/router/middleware"
)

type ServerContext = context.ServerContext

type MiddlewareFunc func(sc ServerContext) echo.MiddlewareFunc
type HandlerFunc = func(sc ServerContext) echo.HandlerFunc

func AddRoutes(e *echo.Echo, sc ServerContext) {
	mountMiddleware(e, sc)
	addFrontendRoutes(e, sc)
	addBackendRoutes(e, sc)
}

func mountMiddleware(e *echo.Echo, sc ServerContext) {
	Use(e, sc, middleware.Session)
}

func addFrontendRoutes(e *echo.Echo, sc ServerContext) {
	GET(e, sc, "/", handler.HandleIndex)
	POST(e, sc, "/logout", handler.HandleLogout)
	GET(e, sc, "/user",
		handler.HandleUser,
		middleware.SessionGuard)
	GET(e, sc, "/market/:id",
		handler.HandleMarket,
		middleware.SessionGuard)
	POST(e, sc, "/market/:id/order",
		handler.HandlePostOrder,
		middleware.SessionGuard,
		middleware.LNDGuard)
	GET(e, sc, "/invoice/:id",
		handler.HandleInvoice,
		middleware.SessionGuard)
}

func addBackendRoutes(e *echo.Echo, sc ServerContext) {
	GET(e, sc, "/api/login", handler.HandleLogin)
	GET(e, sc, "/api/login/callback", handler.HandleLoginCallback)
	GET(e, sc, "/api/session", handler.HandleCheckSession)
	GET(e, sc, "/api/invoice/:id",
		handler.HandleInvoiceStatus,
		middleware.SessionGuard)
}

func GET(e *echo.Echo, sc ServerContext, path string, scF HandlerFunc, scM ...MiddlewareFunc) *echo.Route {
	return e.GET(path, scF(sc), toMiddlewareFunc(sc, scM...)...)
}

func POST(e *echo.Echo, sc ServerContext, path string, scF HandlerFunc, scM ...MiddlewareFunc) *echo.Route {
	return e.POST(path, scF(sc), toMiddlewareFunc(sc, scM...)...)
}

func Use(e *echo.Echo, sc ServerContext, scM ...MiddlewareFunc) {
	e.Use(toMiddlewareFunc(sc, scM...)...)
}

func toMiddlewareFunc(sc ServerContext, scM ...MiddlewareFunc) []echo.MiddlewareFunc {
	var m []echo.MiddlewareFunc
	for _, m_ := range scM {
		m = append(m, m_(sc))
	}
	return m
}
