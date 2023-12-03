package handler

import (
	context_ "context"
	"database/sql"
	"net/http"
	"strings"
	"time"

	"git.ekzyis.com/ekzyis/delphi.market/db"
	"git.ekzyis.com/ekzyis/delphi.market/server/router/context"
	"github.com/labstack/echo/v4"
	"github.com/lightningnetwork/lnd/zpay32"
)

func HandleWithdrawal(sc context.ServerContext) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			w   db.Withdrawal
			u   db.User
			inv *zpay32.Invoice
			tx  *sql.Tx
			err error
		)
		if err := c.Bind(&w); err != nil {
			code := http.StatusBadRequest
			return c.JSON(code, map[string]any{"status": code, "reason": "bolt11 required"})
		}

		if inv, err = zpay32.Decode(w.Bolt11, sc.Lnd.ChainParams); err != nil {
			code := http.StatusBadRequest
			return c.JSON(code, map[string]any{"status": code, "reason": "zpay32 decode error"})
		}

		// transaction start
		ctx, cancel := context_.WithTimeout(c.Request().Context(), 5*time.Second)
		defer cancel()
		if tx, err = sc.Db.BeginTx(ctx, nil); err != nil {
			return err
		}
		defer tx.Commit()

		u = c.Get("session").(db.User)
		w.Pubkey = u.Pubkey

		// TODO deduct network fee from user balance
		if u.Msats < int64(*inv.MilliSat) {
			tx.Rollback()
			code := http.StatusBadRequest
			return c.JSON(code, map[string]any{"status": code, "reason": "insufficient balance"})
		}

		// create withdrawal
		if err = sc.Db.CreateWithdrawal(tx, ctx, &w); err != nil {
			tx.Rollback()
			if strings.Contains(err.Error(), "violates unique constraint") {
				code := http.StatusBadRequest
				return c.JSON(code, map[string]any{"status": code, "reason": "bolt11 already submitted"})
			}
			return err
		}

		// pay invoice via LND
		if err = sc.Lnd.PayInvoice(tx, w.Bolt11); err != nil {
			tx.Rollback()
			return err
		}

		// deduct balance from user
		if _, err = tx.ExecContext(ctx, "UPDATE users SET msats = msats - $1 WHERE pubkey = $2", int64(*inv.MilliSat), u.Pubkey); err != nil {
			tx.Rollback()
			return err
		}

		tx.Commit()

		return c.JSON(http.StatusOK, nil)
	}
}
