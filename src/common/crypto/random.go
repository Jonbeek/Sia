package crypto

import (
	"crypto/rand"
)

// This function gets its own file because I like to have
// the flexibility in deciding to change which random number
// generator to use.
func RandomByteSlice(numBytes int) (randomBytes []byte, err error) {
	randomBytes = make([]byte, numBytes)
	rand.Read(randomBytes)
	return
}
