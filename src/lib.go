package main

import (
	"encoding/base64"

	"github.com/skip2/go-qrcode"
)

func Max(x, y int) int {
	if x < y {
		return y
	}
	return x
}

func ToQR(s string) (string, error) {
	png, err := qrcode.Encode(s, qrcode.Medium, 256)
	if err != nil {
		return "", err
	}
	qr := base64.StdEncoding.EncodeToString([]byte(png))
	return qr, nil
}
