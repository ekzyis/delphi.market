package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/lightninglabs/lndclient"
	"github.com/lightningnetwork/lnd/lnrpc/invoicesrpc"
	"github.com/lightningnetwork/lnd/lntypes"
	"github.com/lightningnetwork/lnd/lnwire"
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
	lndclient.GrpcLndServices
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
	rpcLndServices, err := lndclient.NewLndServices(&lndclient.LndServicesConfig{
		LndAddress:  LndHost,
		MacaroonDir: LndMacaroonDir,
		TLSPath:     LndCert,
		Network:     lndclient.NetworkRegtest,
	})
	if err != nil {
		log.Println(err)
		return
	}
	lnd = &LndClient{GrpcLndServices: *rpcLndServices}
	ver := lnd.Version
	log.Printf("Connected to %s running LND v%s", LndHost, ver.Version)
	lndEnabled = true

}

func lndGuard(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if lndEnabled {
			return next(c)
		}
		return serveError(c, 405)
	}
}

func (lnd *LndClient) GenerateNewPreimage() (lntypes.Preimage, error) {
	randomBytes := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, randomBytes)
	if err != nil {
		return lntypes.Preimage{}, err
	}
	preimage, err := lntypes.MakePreimage(randomBytes)
	if err != nil {
		return lntypes.Preimage{}, err
	}
	return preimage, nil
}

func (lnd *LndClient) CreateInvoice(pubkey string, msats int) (*Invoice, error) {
	expiry := time.Hour
	preimage, err := lnd.GenerateNewPreimage()
	if err != nil {
		return nil, err
	}
	hash := preimage.Hash()
	paymentRequest, err := lnd.Invoices.AddHoldInvoice(context.TODO(), &invoicesrpc.AddInvoiceData{
		Hash:   &hash,
		Value:  lnwire.MilliSatoshi(msats),
		Expiry: int64(expiry),
	})
	if err != nil {
		return nil, err
	}
	lnInvoice, err := lnd.Client.LookupInvoice(context.TODO(), hash)
	if err != nil {
		return nil, err
	}
	dbInvoice := Invoice{
		Session:        Session{pubkey},
		Msats:          msats,
		Preimage:       preimage.String(),
		PaymentRequest: paymentRequest,
		PaymentHash:    hash.String(),
		CreatedAt:      lnInvoice.CreationDate,
		ExpiresAt:      lnInvoice.CreationDate.Add(expiry),
	}
	if err := db.CreateInvoice(&dbInvoice); err != nil {
		return nil, err
	}
	return &dbInvoice, nil
}

func (lnd *LndClient) CheckInvoice(hash lntypes.Hash) {
	if !lndEnabled {
		log.Printf("LND disabled, skipping checking invoice: hash=%s", hash)
		return
	}

	var invoice Invoice
	if err := db.FetchInvoice(&FetchInvoiceWhere{Hash: hash.String()}, &invoice); err != nil {
		log.Println(err)
		return
	}

	loopPause := 5 * time.Second
	handleLoopError := func(err error) {
		log.Println(err)
		time.Sleep(loopPause)
	}

	for {
		log.Printf("lookup invoice: hash=%s", hash)
		lnInvoice, err := lnd.Client.LookupInvoice(context.TODO(), hash)
		if err != nil {
			handleLoopError(err)
			continue
		}
		if time.Now().After(invoice.ExpiresAt) {
			if err := lnd.Invoices.CancelInvoice(context.TODO(), hash); err != nil {
				handleLoopError(err)
				continue
			}
			log.Printf("invoice expired: hash=%s", hash)
			break
		}
		if lnInvoice.AmountPaid > 0 {
			preimage, err := lntypes.MakePreimageFromStr(invoice.Preimage)
			if err != nil {
				handleLoopError(err)
				continue
			}
			// TODO settle invoice after matching order was found
			if err := lnd.Invoices.SettleInvoice(context.TODO(), preimage); err != nil {
				handleLoopError(err)
				continue
			}
			if err := db.ConfirmInvoice(hash.String(), time.Now(), int(lnInvoice.AmountPaid)); err != nil {
				handleLoopError(err)
				continue
			}
			log.Printf("invoice confirmed: hash=%s", hash)
			break
		}
		time.Sleep(loopPause)
	}
}

func invoice(c echo.Context) error {
	invoiceId := c.Param("id")
	var invoice Invoice
	if err := db.FetchInvoice(&FetchInvoiceWhere{Id: invoiceId}, &invoice); err == sql.ErrNoRows {
		return echo.NewHTTPError(http.StatusNotFound, "Not Found")
	} else if err != nil {
		return err
	}
	session := c.Get("session").(Session)
	if invoice.Pubkey != session.Pubkey {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}
	hash, err := lntypes.MakeHashFromStr(invoice.PaymentHash)
	if err != nil {
		return err
	}
	go lnd.CheckInvoice(hash)
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
	if err := db.FetchInvoice(&FetchInvoiceWhere{Id: invoiceId}, &invoice); err == sql.ErrNoRows {
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
