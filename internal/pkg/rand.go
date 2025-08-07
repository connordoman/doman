package pkg

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
)

func RandomBytes(length int) ([]byte, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func RandomString(length int) (string, error) {
	b, err := RandomBytes(length)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func randBase64(length int, urlSafe bool) (string, error) {
	b, err := RandomBytes(length)
	if err != nil {
		return "", err
	}

	if urlSafe {
		return base64.RawURLEncoding.EncodeToString(b), nil
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func RandomBase64(length int) (string, error) {
	return randBase64(length, false)
}

func RandomBase64URL(length int) (string, error) {
	return randBase64(length, true)
}

func RandomHex(length int) (string, error) {
	b, err := RandomBytes(length / 2)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}
