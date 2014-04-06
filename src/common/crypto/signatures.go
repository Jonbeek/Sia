// Sia uses libsodium to handle all of the crypto
package crypto

// #cgo LDFLAGS: -lsodium
// #include<sodium.h>
import "C"

import (
	"fmt"
	"unsafe"
)

// Needs no input, produces 2 strings and an int as output
// first string is a public key, second is a secret key
// the int is the libsodium return value
func createKeyPair() (publicKey string, secretKey string, err error) {
	// Create byte arrays for C to modify
	// Sizes are hardcopied from constants in libsodium
	publicKeyBytes := make([]byte, 32)
	secretKeyBytes := make([]byte, 64)

	// Initialize libsodium and create keys
	C.sodium_init()
	returnValue := C.crypto_sign_keypair((*C.uchar)(unsafe.Pointer(&publicKeyBytes[0])), (*C.uchar)(unsafe.Pointer(&secretKeyBytes[0])))
	if returnValue != 0 {
		err = fmt.Errorf("Key Creation Failed!")
	}

	publicKey = string(publicKeyBytes)
	secretKey = string(secretKeyBytes)
	return
}
