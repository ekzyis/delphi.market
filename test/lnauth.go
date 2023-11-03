package test

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

func Sign(k1_ string) (string, string, error) {
	var (
		sk  *secp256k1.PrivateKey
		k1  []byte
		sig []byte
		err error
	)
	if k1, err = hex.DecodeString(k1_); err != nil {
		return "", "", err
	}
	if sk, err = secp256k1.GeneratePrivateKey(); err != nil {
		return "", "", err
	}
	if sig, err = ecdsa.SignASN1(rand.Reader, sk.ToECDSA(), k1); err != nil {
		return "", "", err
	}
	return hex.EncodeToString(sk.PubKey().SerializeCompressed()), hex.EncodeToString(sig), nil
}
