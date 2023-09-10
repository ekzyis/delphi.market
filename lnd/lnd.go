package lnd

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/lightninglabs/lndclient"
	"github.com/namsral/flag"
)

var (
	lnd     *LNDClient
	Enabled bool
)

type LNDClient struct {
	lndclient.GrpcLndServices
}

func init() {
	var (
		lndCert        string
		lndMacaroonDir string
		lndHost        string
		rpcLndServices *lndclient.GrpcLndServices
		err            error
	)
	if err = godotenv.Load(); err != nil {
		log.Fatalf("error loading env vars: %s", err)
	}
	flag.StringVar(&lndCert, "LND_CERT", "", "Path to LND TLS certificate")
	flag.StringVar(&lndMacaroonDir, "LND_MACAROON_DIR", "", "LND macaroon directory")
	flag.StringVar(&lndHost, "LND_HOST", "localhost:10001", "LND gRPC server address")
	flag.Parse()
	if rpcLndServices, err = lndclient.NewLndServices(&lndclient.LndServicesConfig{
		LndAddress:  lndHost,
		MacaroonDir: lndMacaroonDir,
		TLSPath:     lndCert,
		// TODO: make network configurable
		Network: lndclient.NetworkRegtest,
	}); err != nil {
		log.Println(err)
		Enabled = false
		return
	}
	lnd = &LNDClient{GrpcLndServices: *rpcLndServices}
	log.Printf("Connected to %s running LND v%s", lndHost, lnd.Version.Version)
	Enabled = true
}
