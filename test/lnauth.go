package test

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

func GenerateKeyPair() (*secp256k1.PrivateKey, *secp256k1.PublicKey, error) {
	var (
		sk  *secp256k1.PrivateKey
		err error
	)
	if sk, err = secp256k1.GeneratePrivateKey(); err != nil {
		return nil, nil, err
	}
	return sk, sk.PubKey(), nil
}

func Sign(sk *secp256k1.PrivateKey, k1_ string) (string, error) {
	var (
		k1  []byte
		sig []byte
		err error
	)
	if k1, err = hex.DecodeString(k1_); err != nil {
		return "", err
	}
	if sig, err = ecdsa.SignASN1(rand.Reader, sk.ToECDSA(), k1); err != nil {
		return "", err
	}
	return hex.EncodeToString(sig), nil
}
