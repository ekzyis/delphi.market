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
		CreatedAt time.Time
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
		Id          Serial    `json:"id"`
		Description string    `json:"description"`
		EndDate     time.Time `json:"endDate"`
		InvoiceId   UUID
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
		Description    string
		Status         string
	}
	Order struct {
		Id               UUID
		CreatedAt        time.Time
		DeletedAt        null.Time
		ShareId          string `json:"sid"`
		ShareDescription string
		Share
		Pubkey    string
		Side      string `json:"side"`
		Quantity  int64  `json:"quantity"`
		Price     int64  `json:"price"`
		InvoiceId null.String
		Invoice
		OrderId UUID
	}
)
