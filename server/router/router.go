package router

import (
	"github.com/labstack/echo/v4"

	"git.ekzyis.com/ekzyis/delphi.market/env"
	"git.ekzyis.com/ekzyis/delphi.market/server/router/handler"
	"git.ekzyis.com/ekzyis/delphi.market/server/router/middleware"
)

func AddRoutes(e *echo.Echo) {
	envVars := map[string]any{
		"PUBLIC_URL":       env.PublicURL,
		"COMMIT_SHORT_SHA": env.CommitShortSha,
		"COMMIT_LONG_SHA":  env.CommitLongSha,
		"VERSION":          env.Version,
	}
	e.Use(middleware.Session(envVars))
	e.GET("/", handler.HandleIndex(envVars))
	e.GET("/login", handler.HandleLogin(envVars))
	e.GET("/api/login", handler.HandleLoginCallback(envVars))
	e.GET("/api/session", handler.HandleCheckSession(envVars))
	e.POST("/logout", handler.HandleLogout(envVars))
	e.GET("/user",
		handler.HandleUser(envVars),
		middleware.SessionGuard(envVars))
	e.GET("/market/:id",
		handler.HandleMarket(envVars),
		middleware.SessionGuard(envVars))
	e.POST("/market/:id/order",
		handler.HandlePostOrder(envVars),
		middleware.SessionGuard(envVars),
		middleware.LNDGuard(envVars))
	e.GET("/invoice/:id",
		handler.HandleInvoice(envVars),
		middleware.SessionGuard(envVars),
		middleware.LNDGuard(envVars),
	)
	e.GET("/api/invoice/:id",
		handler.HandleInvoiceAPI(envVars),
		middleware.SessionGuard(envVars),
		middleware.LNDGuard(envVars),
	)
}
