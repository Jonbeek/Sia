package crypto

// #include <sodium.h>
import "C"

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

// This function generates a random int [0, ceiling]
func RandomInt(ceiling int) int {
	return int(C.randombytes_uniform(C.uint32_t(ceiling)))
}
