package main

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcutil/bech32"
)

type LnAuth struct {
	k1    string
	lnurl string
}

type LnAuthResponse struct {
	K1  string `query:"k1"`
	Sig string `query:"sig"`
	Key string `query:"key"`
}

type Session struct {
	pubkey string
}

func lnAuth() (*LnAuth, error) {
	k1 := make([]byte, 32)
	_, err := rand.Read(k1)
	if err != nil {
		return nil, fmt.Errorf("rand.Read error: %w", err)
	}
	k1hex := hex.EncodeToString(k1)
	url := []byte(fmt.Sprintf("https://%s/api/login?tag=login&k1=%s&action=login", PUBLIC_URL, k1hex))
	conv, err := bech32.ConvertBits(url, 8, 5, true)
	if err != nil {
		return nil, fmt.Errorf("bech32.ConvertBits error: %w", err)
	}
	lnurl, err := bech32.Encode("lnurl", conv)
	if err != nil {
		return nil, fmt.Errorf("bech32.Encode error: %w", err)
	}
	return &LnAuth{k1hex, lnurl}, nil
}

func lnAuthVerify(r *LnAuthResponse) (bool, error) {
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
