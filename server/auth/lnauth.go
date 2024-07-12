package auth

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcutil/bech32"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"

	"git.ekzyis.com/ekzyis/delphi.market/env"
)

type LnAuth struct {
	K1    string
	LNURL string
}

type LnAuthCallback struct {
	K1     string `query:"k1"`
	Sig    string `query:"sig"`
	Key    string `query:"key"`
	Action string `query:"action"`
}

func NewLnAuth(action string) (*LnAuth, error) {
	var (
		k1        = make([]byte, 32)
		k1hex     string
		url       []byte
		bech32Url []byte
		lnurl     string
		err       error
	)

	if _, err := rand.Read(k1); err != nil {
		return nil, fmt.Errorf("rand.Read error: %w", err)
	}

	k1hex = hex.EncodeToString(k1)
	url = []byte(fmt.Sprintf("https://%s/api/lnauth/callback?tag=login&k1=%s&action=%s", env.PublicURL, k1hex, action))

	if bech32Url, err = bech32.ConvertBits(url, 8, 5, true); err != nil {
		return nil, fmt.Errorf("bech32.ConvertBits error: %w", err)
	}

	if lnurl, err = bech32.Encode("lnurl", bech32Url); err != nil {
		return nil, fmt.Errorf("bech32.Encode error: %w", err)
	}

	return &LnAuth{k1hex, lnurl}, nil
}

func VerifyLNAuth(r *LnAuthCallback) (bool, error) {
	var (
		k1Bytes, sigBytes, keyBytes []byte
		key                         *secp256k1.PublicKey
		err                         error
	)

	if k1Bytes, err = hex.DecodeString(r.K1); err != nil {
		return false, fmt.Errorf("k1 decode error: %w", err)
	}

	if sigBytes, err = hex.DecodeString(r.Sig); err != nil {
		return false, fmt.Errorf("sig decode error: %w", err)
	}

	if keyBytes, err = hex.DecodeString(r.Key); err != nil {
		return false, fmt.Errorf("key decode error: %w", err)
	}

	if key, err = btcec.ParsePubKey(keyBytes); err != nil {
		return false, fmt.Errorf("key parse error: %w", err)
	}

	ecdsaKey := ecdsa.PublicKey{Curve: btcec.S256(), X: key.X(), Y: key.Y()}
	return ecdsa.VerifyASN1(&ecdsaKey, k1Bytes, sigBytes), nil
}
