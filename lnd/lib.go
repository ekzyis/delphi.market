package lnd

import (
	"crypto/rand"
	"io"

	"github.com/lightningnetwork/lnd/lntypes"
)

func generateNewPreimage() (lntypes.Preimage, error) {
	randomBytes := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, randomBytes)
	if err != nil {
		return lntypes.Preimage{}, err
	}
	preimage, err := lntypes.MakePreimage(randomBytes)
	if err != nil {
		return lntypes.Preimage{}, err
	}
	return preimage, nil
}
