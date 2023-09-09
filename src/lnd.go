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
	lndEnabled     bool
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
	lndEnabled = false
	rpcClient, err := lndclient.NewBasicClient(LndHost, LndCert, LndMacaroonDir, "regtest")
	if err != nil {
		log.Println(err)
		return
	}
	lnd = &LndClient{LightningClient: rpcClient}
	if info, err := lnd.GetInfo(context.TODO(), &lnrpc.GetInfoRequest{}); err != nil {
		log.Printf("LND connection error: %v\n", err)
		return
	} else {
		version := strings.Split(info.Version, " ")[0]
		log.Printf("Connected to %s running LND v%s", LndHost, version)
		lndEnabled = true
	}
}

func lndGuard(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if lndEnabled {
			return next(c)
		}
		return serveError(c, 405)
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
	if !lndEnabled {
		log.Printf("LND disabled, skipping checking invoice: hash=%s", hash)
		return
	}
	for {
		log.Printf("lookup invoice: hash=%s", hash)
		invoice, err := lnd.LookupInvoice(context.TODO(), &lnrpc.PaymentHash{RHashStr: hash})
		if err != nil {
			log.Println(err)
			time.Sleep(5 * time.Second)
			continue
		}
		if time.Now().After(time.Unix(invoice.CreationDate+invoice.Expiry, 0)) {
			log.Printf("invoice expired: hash=%s", hash)
			break
		}
		if invoice.SettleDate != 0 && invoice.AmtPaidMsat > 0 {
			if err := db.ConfirmInvoice(hash, time.Unix(invoice.SettleDate, 0), int(invoice.AmtPaidMsat)); err != nil {
				log.Println(err)
				time.Sleep(5 * time.Second)
				continue
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
	go lnd.CheckInvoice(invoice.PaymentHash)
	qr, err := ToQR(invoice.PaymentRequest)
	if err != nil {
		return err
	}
	status := ""
	if invoice.ConfirmedAt.Valid {
		status = "Paid"
	} else if time.Now().After(invoice.ExpiresAt) {
		status = "Expired"
	}
	data := map[string]any{
		"session": c.Get("session"),
		"Invoice": invoice,
		"Status":  status,
		"lnurl":   invoice.PaymentRequest,
		"qr":      qr,
	}
	return c.Render(http.StatusOK, "invoice.html", data)
}

func invoiceStatus(c echo.Context) error {
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
	invoice.Preimage = ""
	return c.JSON(http.StatusOK, invoice)
}
