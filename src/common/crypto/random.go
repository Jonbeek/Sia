package crypto

import (
	"crypto/rand"
	"fmt"
)

func Random(numBytes int) (entropy []byte, err error) {
	entropy = make([]byte, numBytes)
	bytesCopied, err := rand.Read(entropy)

	if err != nil {
		return
	}

	if bytesCopied != numBytes {
		err = fmt.Errorf("Expected %v bytes, created %v bytes", numBytes, bytesCopied)
	}

	return
}
