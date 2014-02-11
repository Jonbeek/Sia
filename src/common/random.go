package common

import (
	"crypto/rand"
)

func RandomString(n uint) (random string, err error) {
	r := make([]byte, n)
	_, err = rand.Read(r)
	random = string(r)
	return
}
