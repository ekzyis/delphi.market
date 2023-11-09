package handler

import (
	"database/sql"
	"net/http"
	"time"

	"git.ekzyis.com/ekzyis/delphi.market/db"
	"git.ekzyis.com/ekzyis/delphi.market/lib"
	"git.ekzyis.com/ekzyis/delphi.market/server/router/context"
	"github.com/labstack/echo/v4"
	"github.com/lightningnetwork/lnd/lntypes"
)

func HandleInvoiceStatus(sc context.ServerContext) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			invoiceId string
			invoice   db.Invoice
			u         db.User
			qr        string
			err       error
		)
		invoiceId = c.Param("id")
		if err = sc.Db.FetchInvoice(&db.FetchInvoiceWhere{Id: invoiceId}, &invoice); err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound)
		} else if err != nil {
			return err
		}
		if u = c.Get("session").(db.User); invoice.Pubkey != u.Pubkey {
			return echo.NewHTTPError(http.StatusUnauthorized)
		}
		if qr, err = lib.ToQR(invoice.PaymentRequest); err != nil {
			return err
		}
		invoice.Preimage = ""
		data := map[string]any{
			"Id":             invoice.Id,
			"Msats":          invoice.Msats,
			"MsatsReceived":  invoice.MsatsReceived,
			"Hash":           invoice.Hash,
			"PaymentRequest": invoice.PaymentRequest,
			"CreatedAt":      invoice.CreatedAt,
			"ExpiresAt":      invoice.ExpiresAt,
			"ConfirmedAt":    invoice.ConfirmedAt,
			"HeldSince":      invoice.HeldSince,
			"Description":    invoice.Description,
			"Qr":             qr,
		}
		return c.JSON(http.StatusOK, data)
	}
}

func HandleInvoice(sc context.ServerContext) echo.HandlerFunc {
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
		if err = sc.Db.FetchInvoice(&db.FetchInvoiceWhere{Id: invoiceId}, &invoice); err == sql.ErrNoRows {
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
		go sc.Lnd.CheckInvoice(sc.Db, hash)
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
		return sc.Render(c, http.StatusOK, "invoice.html", data)
	}
}
