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
	GET(e, sc, "/user",
		handler.HandleUser,
		middleware.SessionGuard)
	GET(e, sc, "/market/:id",
		handler.HandleMarket,
		middleware.SessionGuard)
	POST(e, sc, "/market/:id/order",
		handler.HandleOrder,
		middleware.SessionGuard,
		middleware.LNDGuard)
	GET(e, sc, "/invoice/:id",
		handler.HandleInvoice,
		middleware.SessionGuard)
}

func addBackendRoutes(e *echo.Echo, sc ServerContext) {
	GET(e, sc, "/api/markets", handler.HandleMarkets)
	POST(e, sc, "/api/market",
		handler.HandleCreateMarket,
		middleware.SessionGuard,
		middleware.LNDGuard)
	GET(e, sc, "/api/market/:id", handler.HandleMarket)
	GET(e, sc, "/api/market/:id/orders", handler.HandleMarketOrders)
	GET(e, sc, "/api/market/:id/stats", handler.HandleMarketStats)
	POST(e, sc, "/api/order",
		handler.HandleOrder,
		middleware.SessionGuard,
		middleware.LNDGuard)
	DELETE(e, sc, "/api/order/:id",
		handler.HandleDeleteOrder,
		middleware.SessionGuard)
	GET(e, sc, "/api/orders",
		handler.HandleOrders,
		middleware.SessionGuard)
	GET(e, sc, "/api/login", handler.HandleLogin)
	GET(e, sc, "/api/login/callback", handler.HandleLoginCallback)
	POST(e, sc, "/api/logout", handler.HandleLogout)
	GET(e, sc, "/api/session", handler.HandleCheckSession)
	GET(e, sc, "/api/invoice/:id",
		handler.HandleInvoiceStatus,
		middleware.SessionGuard)
	GET(e, sc, "/api/invoices",
		handler.HandleInvoices,
		middleware.SessionGuard)
	POST(e, sc, "/api/withdrawal",
		handler.HandleWithdrawal,
		middleware.SessionGuard,
		middleware.LNDGuard)
}

func GET(e *echo.Echo, sc ServerContext, path string, scF HandlerFunc, scM ...MiddlewareFunc) *echo.Route {
	return e.GET(path, scF(sc), toMiddlewareFunc(sc, scM...)...)
}

func POST(e *echo.Echo, sc ServerContext, path string, scF HandlerFunc, scM ...MiddlewareFunc) *echo.Route {
	return e.POST(path, scF(sc), toMiddlewareFunc(sc, scM...)...)
}

func DELETE(e *echo.Echo, sc ServerContext, path string, scF HandlerFunc, scM ...MiddlewareFunc) *echo.Route {
	return e.DELETE(path, scF(sc), toMiddlewareFunc(sc, scM...)...)
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
