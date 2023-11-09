package main

import (
	"fmt"
	"log"
	"net/http"

	"git.ekzyis.com/ekzyis/delphi.market/db"
	"git.ekzyis.com/ekzyis/delphi.market/env"
	"git.ekzyis.com/ekzyis/delphi.market/lnd"
	"git.ekzyis.com/ekzyis/delphi.market/server"
	"git.ekzyis.com/ekzyis/delphi.market/server/router"
	"github.com/lightninglabs/lndclient"
	"github.com/namsral/flag"
)

var (
	s *server.Server
)

func init() {
	var (
		dbUrl          string
		lndAddress     string
		lndCert        string
		lndMacaroonDir string
		db_            *db.DB
		lnd_           *lnd.LNDClient
		ctx            router.ServerContext
		err            error
	)
	if err = env.Load(); err != nil {
		log.Fatalf("error loading env vars: %v", err)
	}
	flag.StringVar(&dbUrl, "DATABASE_URL", "delphi.market", "Public URL of website")
	flag.StringVar(&lndAddress, "LND_ADDRESS", "localhost:10001", "LND gRPC server address")
	flag.StringVar(&lndCert, "LND_CERT", "", "Path to LND TLS certificate")
	flag.StringVar(&lndMacaroonDir, "LND_MACAROON_DIR", "", "LND macaroon directory")
	env.Parse()
	figlet()
	log.Printf("Commit:      %s", env.CommitShortSha)
	log.Printf("Public URL:  %s", env.PublicURL)
	log.Printf("Environment: %s", env.Env)

	if db_, err = db.New(dbUrl); err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}

	if lnd_, err = lnd.New(&lnd.LNDConfig{
		LndAddress:  lndAddress,
		TLSPath:     lndCert,
		MacaroonDir: lndMacaroonDir,
		Network:     lndclient.NetworkRegtest,
	}); err != nil {
		log.Printf("[warn] error connecting to LND: %v\n", err)
		lnd_ = nil
	} else {
		lnd_.CheckInvoices(db_)
	}

	ctx = server.ServerContext{
		PublicURL:      env.PublicURL,
		CommitShortSha: env.CommitShortSha,
		CommitLongSha:  env.CommitLongSha,
		Version:        env.Version,
		Db:             db_,
		Lnd:            lnd_,
	}
	s = server.New(ctx)
}

func figlet() {
	log.Println(
		"\n" +
			"     _      _       _     _ \n" +
			"  __| | ___| |_ __ | |__ (_)\n" +
			" / _` |/ _ \\ | '_ \\| '_ \\| |\n" +
			"| (_| |  __/ | |_) | | | | |\n" +
			" \\__,_|\\___|_| .__/|_| |_|_|\n" +
			"             |_| .market       \n" +
			"----------------------------",
	)
}

func main() {
	if err := s.Start(fmt.Sprintf("%s:%d", "127.0.0.1", env.Port)); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
