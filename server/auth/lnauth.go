package auth

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcutil/bech32"

	"git.ekzyis.com/ekzyis/delphi.market/env"
)

type LNAuth struct {
	K1    string
	LNURL string
}

type LNAuthResponse struct {
	K1  string `query:"k1"`
	Sig string `query:"sig"`
	Key string `query:"key"`
}

func NewLNAuth() (*LNAuth, error) {
	k1 := make([]byte, 32)
	_, err := rand.Read(k1)
	if err != nil {
		return nil, fmt.Errorf("rand.Read error: %w", err)
	}
	k1hex := hex.EncodeToString(k1)
	url := []byte(fmt.Sprintf("https://%s/api/login/callback?tag=login&k1=%s&action=login", env.PublicURL, k1hex))
	conv, err := bech32.ConvertBits(url, 8, 5, true)
	if err != nil {
		return nil, fmt.Errorf("bech32.ConvertBits error: %w", err)
	}
	lnurl, err := bech32.Encode("lnurl", conv)
	if err != nil {
		return nil, fmt.Errorf("bech32.Encode error: %w", err)
	}
	return &LNAuth{k1hex, lnurl}, nil
}

func VerifyLNAuth(r *LNAuthResponse) (bool, error) {
	var k1Bytes, sigBytes, keyBytes []byte
	k1Bytes, err := hex.DecodeString(r.K1)
	if err != nil {
		return false, fmt.Errorf("k1 decode error: %w", err)
	}
	sigBytes, err = hex.DecodeString(r.Sig)
	if err != nil {
		return false, fmt.Errorf("sig decode error: %w", err)
	}
	keyBytes, err = hex.DecodeString(r.Key)
	if err != nil {
		return false, fmt.Errorf("key decode error: %w", err)
	}
	key, err := btcec.ParsePubKey(keyBytes)
	if err != nil {
		return false, fmt.Errorf("key parse error: %w", err)
	}
	ecdsaKey := ecdsa.PublicKey{Curve: btcec.S256(), X: key.X(), Y: key.Y()}
	return ecdsa.VerifyASN1(&ecdsaKey, k1Bytes, sigBytes), nil
}
