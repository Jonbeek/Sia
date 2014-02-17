package common

import (
	"crypto/rand"
	"encoding/hex"
)

func RandomString(n uint) (random string, err error) {
	r := make([]byte, n)
	_, err = rand.Read(r)
	random = hex.EncodeToString(r)
	return
}
