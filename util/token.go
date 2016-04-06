package util

import (
	"crypto/rand"
	"encoding/base32"
)

// RandomToken returns a cryptographically strong random token string.
// The Token is generated from 512 random bits and encoded via base32.StdEncoding
func RandomToken() string {
	rnd := make([]byte, 64)

	_, err := rand.Read(rnd)
	if err != nil {
		panic(err)
	}

	return base32.StdEncoding.EncodeToString(rnd)
}
