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
func CreateKeyPair() (publicKey string, secretKey string, err error) {
	// Create byte arrays for C to modify
	// Sizes are hardcopied from constants in libsodium
	publicKeyBytes := make([]byte, PublicKeySize)
	secretKeyBytes := make([]byte, SecretKeySize)

	// Initialize libsodium and create keys
	C.sodium_init()
	errorCode := C.crypto_sign_keypair((*C.uchar)(unsafe.Pointer(&publicKeyBytes[0])), (*C.uchar)(unsafe.Pointer(&secretKeyBytes[0])))

	// Check that the function returned without error
	if errorCode != 0 {
		err = fmt.Errorf("Key Creation Failed!")
		return
	}

	publicKey = string(publicKeyBytes)
	secretKey = string(secretKeyBytes)
	return
}

func Sign(secretKey string, message string) (signature string, err error) {
	// Make sure secretKey is the correct number of bytes
	if len(secretKey) != SecretKeySize {
		err = fmt.Errorf("Secret Key is incorrect size!")
		return
	}

	signatureBytes := make([]byte, len(message)+SignatureSize)
	var signatureLen uint64
	messageBytes := []byte(message)
	secretKeyBytes := []byte(secretKey)

	errorCode := C.crypto_sign((*C.uchar)(unsafe.Pointer(&signatureBytes[0])), (*C.ulonglong)(unsafe.Pointer(&signatureLen)), (*C.uchar)(unsafe.Pointer(&messageBytes[0])), C.ulonglong(len(message)), (*C.uchar)(unsafe.Pointer(&secretKeyBytes[0])))

	// Check that the function returned without error
	if errorCode != 0 {
		err = fmt.Errorf("Signature Failed!")
		return
	}

	signature = string(signatureBytes)
	return
}

func Verify(verificationKey string, signedMessage string) (verified bool, err error) {
	// Check length of verification key
	if len(verificationKey) != PublicKeySize {
		err = fmt.Errorf("Public Key is incorrect size!")
		return
	}

	messageBytes := make([]byte, len(signedMessage)-SignatureSize)
	var messageLen uint64
	signedMessageBytes := []byte(signedMessage)
	verificationKeyBytes := []byte(verificationKey)

	success := C.crypto_sign_open((*C.uchar)(unsafe.Pointer(&messageBytes[0])), (*C.ulonglong)(unsafe.Pointer(&messageLen)), (*C.uchar)(unsafe.Pointer(&signedMessageBytes[0])), C.ulonglong(len(signedMessage)), (*C.uchar)(unsafe.Pointer(&verificationKeyBytes[0])))

	if success == 0 {
		verified = true
		return
	}

	verified = false
	return
}
