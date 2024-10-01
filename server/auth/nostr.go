package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"
)

type NostrAuth struct {
	K1 string
}

type NostrAuthCallback struct {
	K1        string `query:"k1"`
	Sig       string `query:"sig"`
	PubKey    string `query:"key"`
	Action    string `query:"action"`
	CreatedAt int    `query:"created_at"`
}

// reference: https://github.com/nbd-wtf/go-nostr
type Event struct {
	ID        string
	PubKey    string
	CreatedAt int
	Kind      int
	Tags      Tags
	Content   string
	Sig       string

	// anything here will be mashed together with the main event object when serializing
	// extra map[string]any
}

type Tag []string
type Tags []Tag

func NewNostrAuth(action string) (*NostrAuth, error) {
	var (
		k1  = make([]byte, 32)
		err error
	)

	if _, err = rand.Read(k1); err != nil {
		return nil, fmt.Errorf("rand.Read error: %w", err)
	}

	return &NostrAuth{K1: hex.EncodeToString(k1)}, nil
}

func VerifyNostrAuth(r *NostrAuthCallback) (bool, error) {
	// https://github.com/nbd-wtf/go-nostr/blob/master/signature.go

	e := Event{
		PubKey:    r.PubKey,
		Sig:       r.Sig,
		Kind:      22242,
		Content:   "delphi.market authentication",
		Tags:      []Tag{{"challenge", r.K1}},
		CreatedAt: r.CreatedAt,
	}

	return e.CheckSignature()
}

func (tag Tag) marshalTo(dst []byte) []byte {
	dst = append(dst, '[')
	for i, s := range tag {
		if i > 0 {
			dst = append(dst, ',')
		}
		dst = escapeString(dst, s)
	}
	dst = append(dst, ']')
	return dst
}

func (tags Tags) marshalTo(dst []byte) []byte {
	dst = append(dst, '[')
	for i, tag := range tags {
		if i > 0 {
			dst = append(dst, ',')
		}
		dst = tag.marshalTo(dst)
	}
	dst = append(dst, ']')
	return dst
}

func (evt *Event) Serialize() []byte {
	// the serialization process is just putting everything into a JSON array
	// so the order is kept. See NIP-01
	dst := make([]byte, 0)

	// the header portion is easy to serialize
	// [0,"pubkey",created_at,kind,[
	dst = append(dst, []byte(
		fmt.Sprintf(
			"[0,\"%s\",%d,%d,",
			evt.PubKey,
			evt.CreatedAt,
			evt.Kind,
		))...)

	// tags
	dst = evt.Tags.marshalTo(dst)
	dst = append(dst, ',')

	// content needs to be escaped in general as it is user generated.
	dst = escapeString(dst, evt.Content)
	dst = append(dst, ']')

	return dst
}

func escapeString(dst []byte, s string) []byte {
	dst = append(dst, '"')
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '"':
			// quotation mark
			dst = append(dst, []byte{'\\', '"'}...)
		case c == '\\':
			// reverse solidus
			dst = append(dst, []byte{'\\', '\\'}...)
		case c >= 0x20:
			// default, rest below are control chars
			dst = append(dst, c)
		case c == 0x08:
			dst = append(dst, []byte{'\\', 'b'}...)
		case c < 0x09:
			dst = append(dst, []byte{'\\', 'u', '0', '0', '0', '0' + c}...)
		case c == 0x09:
			dst = append(dst, []byte{'\\', 't'}...)
		case c == 0x0a:
			dst = append(dst, []byte{'\\', 'n'}...)
		case c == 0x0c:
			dst = append(dst, []byte{'\\', 'f'}...)
		case c == 0x0d:
			dst = append(dst, []byte{'\\', 'r'}...)
		case c < 0x10:
			dst = append(dst, []byte{'\\', 'u', '0', '0', '0', 0x57 + c}...)
		case c < 0x1a:
			dst = append(dst, []byte{'\\', 'u', '0', '0', '1', 0x20 + c}...)
		case c < 0x20:
			dst = append(dst, []byte{'\\', 'u', '0', '0', '1', 0x47 + c}...)
		}
	}
	dst = append(dst, '"')
	return dst
}

func (e *Event) CheckSignature() (bool, error) {
	// read and check pubkey
	pk, err := hex.DecodeString(e.PubKey)
	if err != nil {
		return false, fmt.Errorf("event pubkey '%s' is invalid hex: %w", e.PubKey, err)
	}

	pubkey, err := schnorr.ParsePubKey(pk)
	if err != nil {
		return false, fmt.Errorf("event has invalid pubkey '%s': %w", e.PubKey, err)
	}

	// read signature
	s, err := hex.DecodeString(e.Sig)
	if err != nil {
		return false, fmt.Errorf("signature '%s' is invalid hex: %w", e.Sig, err)
	}
	sig, err := schnorr.ParseSignature(s)
	if err != nil {
		return false, fmt.Errorf("failed to parse signature: %w", err)
	}

	// check signature
	hash := sha256.Sum256(e.Serialize())
	return sig.Verify(hash[:], pubkey), nil
}
