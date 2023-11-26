package lnd

import (
	"log"

	"github.com/lightninglabs/lndclient"
)

type LNDClient struct {
	*lndclient.GrpcLndServices
}

type LNDConfig = lndclient.LndServicesConfig

func New(config *LNDConfig) (*LNDClient, error) {
	var (
		rcpLndServices *lndclient.GrpcLndServices
		lnd            *LNDClient
		err            error
	)
	if rcpLndServices, err = lndclient.NewLndServices(config); err != nil {
		return nil, err
	}
	lnd = &LNDClient{GrpcLndServices: rcpLndServices}
	log.Printf("Connected to %s running LND v%s", config.LndAddress, lnd.Version.Version)
	return lnd, nil
}
