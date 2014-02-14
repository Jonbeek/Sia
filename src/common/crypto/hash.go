package crypto

import (
	"crypto/sha256"
)

// error checking?
func Hash(input []byte) (output []byte) {
	hash := sha256.New()
	hash.Write(input)
	output = hash.Sum(nil)
	return
}
