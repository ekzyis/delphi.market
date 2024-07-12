package types

import "time"

type User struct {
	Id          int
	CreatedAt   time.Time
	LnPubkey    string
	NostrPubkey string
	Msats       int64
}
