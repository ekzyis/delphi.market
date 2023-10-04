package db

import (
	"time"

	"gopkg.in/guregu/null.v4"
)

type (
	Serial = int
	UUID   = string
	LNAuth struct {
		K1        string
		LNURL     string
		CreatdAt  time.Time
		SessionId string
	}
	User struct {
		Pubkey   string
		LastSeen time.Time
	}
	Session struct {
		Pubkey    string
		SessionId string
	}
	Market struct {
		Id          Serial
		Description string
		Active      bool
	}
	Share struct {
		Id          UUID
		MarketId    int
		Description string
	}
	Invoice struct {
		Id             UUID
		Pubkey         string
		Msats          int64
		MsatsReceived  int64
		Preimage       string
		Hash           string
		PaymentRequest string
		CreatedAt      time.Time
		ExpiresAt      time.Time
		ConfirmedAt    null.Time
		HeldSince      null.Time
	}
	Order struct {
		Id        UUID
		CreatedAt time.Time
		ShareId   string `form:"share_id"`
		Share
		Pubkey    string
		Side      string `form:"side"`
		Quantity  int64  `form:"quantity"`
		Price     int64  `form:"price"`
		InvoiceId UUID
		Invoice
	}
)
