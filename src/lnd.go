package main

import (
	"context"
	"database/sql"
	"encoding/hex"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/lightninglabs/lndclient"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/namsral/flag"
)

var (
	LndCert        string
	LndMacaroonDir string
	LndHost        string
	lnd            *LndClient
)

type LndClient struct {
	lnrpc.LightningClient
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}
	flag.StringVar(&LndCert, "LND_CERT", "", "Path to LND TLS certificate")
	flag.StringVar(&LndMacaroonDir, "LND_MACAROON_DIR", "", "LND macaroon directory")
	flag.StringVar(&LndHost, "LND_HOST", "localhost:10001", "LND gRPC server address")
	flag.Parse()
	rpcClient, err := lndclient.NewBasicClient(LndHost, LndCert, LndMacaroonDir, "regtest")
	if err != nil {
		panic(err)
	}
	lnd = &LndClient{LightningClient: rpcClient}
	if info, err := lnd.GetInfo(context.TODO(), &lnrpc.GetInfoRequest{}); err != nil {
		panic(err)
	} else {
		version := strings.Split(info.Version, " ")[0]
		log.Printf("Connected to %s running LND v%s", LndHost, version)
	}
}

func (lnd *LndClient) CreateInvoice(pubkey string, msats int) (*Invoice, error) {
	addInvoiceResponse, err := lnd.AddInvoice(context.TODO(), &lnrpc.Invoice{
		ValueMsat: int64(msats),
		Expiry:    3600,
	})
	if err != nil {
		return nil, err
	}
	lnInvoice, err := lnd.LookupInvoice(context.TODO(), &lnrpc.PaymentHash{RHash: addInvoiceResponse.RHash})
	if err != nil {
		return nil, err
	}
	dbInvoice := Invoice{
		Session:        Session{pubkey},
		Msats:          msats,
		Preimage:       hex.EncodeToString(lnInvoice.RPreimage),
		PaymentRequest: lnInvoice.PaymentRequest,
		PaymentHash:    hex.EncodeToString(lnInvoice.RHash),
		CreatedAt:      time.Unix(lnInvoice.CreationDate, 0),
		ExpiresAt:      time.Unix(lnInvoice.CreationDate+lnInvoice.Expiry, 0),
	}
	if err := db.CreateInvoice(&dbInvoice); err != nil {
		return nil, err
	}
	return &dbInvoice, nil
}

func (lnd *LndClient) CheckInvoice(hash string) {
	for {
		log.Printf("lookup invoice: hash=%s", hash)
		invoice, err := lnd.LookupInvoice(context.TODO(), &lnrpc.PaymentHash{RHashStr: hash})
		if err != nil {
			panic(err)
		}
		if time.Now().After(time.Unix(invoice.CreationDate+invoice.Expiry, 0)) {
			log.Printf("invoice expired: hash=%s", hash)
			break
		}
		if invoice.SettleDate != 0 && invoice.AmtPaidMsat > 0 {
			if err := db.ConfirmInvoice(hash, time.Unix(invoice.SettleDate, 0), int(invoice.AmtPaidMsat)); err != nil {
				panic(err)
			}
			log.Printf("invoice confirmed: hash=%s", hash)
			break
		}
		time.Sleep(5 * time.Second)
	}
}

func invoice(c echo.Context) error {
	invoiceId := c.Param("id")
	var invoice Invoice
	if err := db.FetchInvoice(invoiceId, &invoice); err == sql.ErrNoRows {
		return echo.NewHTTPError(http.StatusNotFound, "Not Found")
	} else if err != nil {
		return err
	}
	session := c.Get("session").(Session)
	if invoice.Pubkey != session.Pubkey {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}
	qr, err := ToQR(invoice.PaymentRequest)
	if err != nil {
		return err
	}
	data := map[string]any{
		"session": c.Get("session"),
		"Invoice": invoice,
		"lnurl":   invoice.PaymentRequest,
		"qr":      qr,
	}
	return c.Render(http.StatusOK, "invoice.html", data)
}
