package context

import (
	"context"

	"git.ekzyis.com/ekzyis/delphi.market/db"
	"git.ekzyis.com/ekzyis/delphi.market/lnd"
	"github.com/labstack/echo/v4"
)

type Context struct {
	context.Context
	Environment    string
	PublicURL      string
	CommitShortSha string
	CommitLongSha  string
	Version        string
	Db             *db.DB
	Lnd            *lnd.LNDClient
}

type RenderContextKey string

var (
	EnvContextKey     RenderContextKey = "env"
	SessionContextKey RenderContextKey = "session"
	CommitContextKey  RenderContextKey = "commit"
	ReqPathContextKey RenderContextKey = "reqPath"
)

func RenderContext(sc Context, c echo.Context) context.Context {
	ctx := c.Request().Context()
	ctx = context.WithValue(ctx, EnvContextKey, sc.Environment)
	ctx = context.WithValue(ctx, SessionContextKey, c.Get("session"))
	ctx = context.WithValue(ctx, CommitContextKey, sc.CommitShortSha)
	ctx = context.WithValue(ctx, ReqPathContextKey, c.Request().URL.Path)
	return ctx
}
