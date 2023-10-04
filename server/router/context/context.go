package context

import (
	"git.ekzyis.com/ekzyis/delphi.market/db"
	"git.ekzyis.com/ekzyis/delphi.market/lnd"
	"github.com/labstack/echo/v4"
)

type ServerContext struct {
	Environment    string
	PublicURL      string
	CommitShortSha string
	CommitLongSha  string
	Version        string
	Db             *db.DB
	Lnd            *lnd.LNDClient
}

func (sc *ServerContext) Render(c echo.Context, code int, name string, data map[string]any) error {
	envVars := map[string]any{
		"PUBLIC_URL":       sc.PublicURL,
		"COMMIT_SHORT_SHA": sc.CommitShortSha,
		"COMMIT_LONG_SHA":  sc.CommitLongSha,
		"VERSION":          sc.Version,
	}
	merge(&data, &envVars)
	return c.Render(code, name, data)
}

func merge[T comparable](target *map[T]any, src *map[T]any) {
	for k, v := range *src {
		(*target)[k] = v
	}
}
