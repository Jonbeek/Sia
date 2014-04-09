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
func CreateKeyPair() (publicKey PublicKey, secretKey SecretKey, err error) {
	// Create keys
	errorCode := C.crypto_sign_keypair((*C.uchar)(unsafe.Pointer(&publicKey[0])), (*C.uchar)(unsafe.Pointer(&secretKey[0])))

	// Check that the function returned without error
	if errorCode != 0 {
		err = fmt.Errorf("Key Creation Failed!")
		return
	}

	return
}

// Take a secret key and a message, and use the secret key to sign the message.
func Sign(secretKey SecretKey, message string) (signedMessage SignedMessage, err error) {
	// create C-friendly objects
	signedMessageBytes := make([]byte, len(message)+SignatureSize) // output of C.crypto_sign
	var signatureLen uint64                                        // output of C.crypto_sign
	messageBytes := []byte(message)                                // input of C.crypto_sign

	// sign message
	errorCode := C.crypto_sign((*C.uchar)(unsafe.Pointer(&signedMessageBytes[0])), (*C.ulonglong)(unsafe.Pointer(&signatureLen)), (*C.uchar)(unsafe.Pointer(&messageBytes[0])), C.ulonglong(len(message)), (*C.uchar)(unsafe.Pointer(&secretKey[0])))

	// Check that the function returned without error
	if errorCode != 0 {
		err = fmt.Errorf("Signature Failed!")
		return
	}

	signedMessage.Message = message
	// copies the last SignatureSize bytes of signedMessageBytes into signedMessage.Signature
	copy(signedMessage.Signature[:], signedMessageBytes[len(message):])

	return
}

// takes as input a public key and a signed message
// returns whether the signature is valid or not
func Verify(verificationKey PublicKey, signedMessage SignedMessage) (verified bool, err error) {
	// create C friendly objects
	messageBytes := make([]byte, len(signedMessage.Message))                                   // output of C.crypto_sign_open
	var messageLen uint64                                                                      // output of C.crypto_sign_open
	signedMessageBytes := append([]byte(signedMessage.Message), signedMessage.Signature[:]...) // input of C.crypto_sign

	// verify signature
	success := C.crypto_sign_open((*C.uchar)(unsafe.Pointer(&messageBytes[0])), (*C.ulonglong)(unsafe.Pointer(&messageLen)), (*C.uchar)(unsafe.Pointer(&signedMessageBytes[0])), C.ulonglong(len(signedMessageBytes)), (*C.uchar)(unsafe.Pointer(&verificationKey[0])))

	if success == 0 {
		verified = true
		return
	}

	verified = false
	return
}
