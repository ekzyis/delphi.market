package types

import (
	"time"

	"gopkg.in/guregu/null.v4"
)

type User struct {
	Id          int
	Name        string
	CreatedAt   time.Time
	LnPubkey    null.String
	NostrPubkey null.String
	Msats       int64
}

type Invoice struct {
	Id            int
	UserId        int
	Msats         int64
	MsatsReceived int64
	Hash          string
	Bolt11        string
	CreatedAt     time.Time
	ExpiresAt     time.Time
	ConfirmedAt   null.Time
	HeldSince     bool
	Description   string
}

type Market struct {
	Id          int
	User        User
	Question    string
	Description string
	CreatedAt   time.Time
	EndDate     time.Time
}

type LMSR struct {
	B  float64
	Q1 int
	Q2 int
}

type MarketQuote struct {
	Outcome    int
	AvgPrice   float64
	TotalPrice float64
	Reward     float64
}

type Point struct {
	X time.Time
	Y float64
}
