package handler

import (
	"database/sql"
	"net/http"
	"time"

	"git.ekzyis.com/ekzyis/delphi.market/db"
	"git.ekzyis.com/ekzyis/delphi.market/lib"
	"git.ekzyis.com/ekzyis/delphi.market/lnd"
	"github.com/labstack/echo/v4"
	"github.com/lightningnetwork/lnd/lntypes"
)

func HandleInvoiceAPI(envVars map[string]any) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			invoiceId string
			invoice   db.Invoice
			u         db.User
			err       error
		)
		invoiceId = c.Param("id")
		if err = db.FetchInvoice(&db.FetchInvoiceWhere{Id: invoiceId}, &invoice); err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound)
		} else if err != nil {
			return err
		}
		if u = c.Get("session").(db.User); invoice.Pubkey != u.Pubkey {
			return echo.NewHTTPError(http.StatusUnauthorized)
		}
		invoice.Preimage = ""
		return c.JSON(http.StatusOK, invoice)
	}
}

func HandleInvoice(envVars map[string]any) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			invoiceId string
			invoice   db.Invoice
			u         db.User
			hash      lntypes.Hash
			qr        string
			status    string
			err       error
		)
		invoiceId = c.Param("id")
		if err = db.FetchInvoice(&db.FetchInvoiceWhere{Id: invoiceId}, &invoice); err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound)
		} else if err != nil {
			return err
		}
		if u = c.Get("session").(db.User); invoice.Pubkey != u.Pubkey {
			return echo.NewHTTPError(http.StatusUnauthorized)
		}
		if hash, err = lntypes.MakeHashFromStr(invoice.Hash); err != nil {
			return err
		}
		go lnd.CheckInvoice(hash)
		if qr, err = lib.ToQR(invoice.PaymentRequest); err != nil {
			return err
		}
		if invoice.ConfirmedAt.Valid {
			status = "Paid"
		} else if time.Now().After(invoice.ExpiresAt) {
			status = "Expired"
		}
		data := map[string]any{
			"session": c.Get("session"),
			"invoice": invoice,
			"status":  status,
			"lnurl":   invoice.PaymentRequest,
			"qr":      qr,
		}
		return c.Render(http.StatusOK, "invoice.html", data)
	}
}
