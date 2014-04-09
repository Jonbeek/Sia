package crypto

// #cgo LDFLAGS: -lsodium
// #include<sodium.h>
import "C"

import (
	"fmt"
	"unsafe"
)

type SignedMessage struct {
	Message   string
	Signature Signature
}

func (sm *SignedMessage) CombinedMessage() (combinedMessage string) {
	combinedMessage = string(append([]byte(sm.Message[:]), sm.Signature[:]...))
	combinedMessage = string(append(sm.Signature[:], []byte(sm.Message[:])...))
	return
}

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
	// Points to a signed message of format signature + message after sigining
	signedMessageBytes := make([]byte, len(message)+SignatureSize)
	signedMessagePointer := (*C.uchar)(unsafe.Pointer(&signedMessageBytes[0]))

	// Points to the length of the signed message after signing
	var signatureLen uint64
	lenPointer := (*C.ulonglong)(unsafe.Pointer(&signatureLen))

	// Pointer to the message
	// can't point to an empty slice, get runtime panic
	var messageBytes []byte
	if len(message) != 0 {
		messageBytes = []byte(message)
	} else {
		messageBytes = make([]byte, 1)
	}
	messagePointer := (*C.uchar)(unsafe.Pointer(&messageBytes[0]))

	// How long the message is
	messageLen := C.ulonglong(len(message))

	// Pointer to the signature
	signaturePointer := (*C.uchar)(unsafe.Pointer(&secretKey[0]))

	// sign message
	errorCode := C.crypto_sign(signedMessagePointer, lenPointer, messagePointer, messageLen, signaturePointer)

	// Check that the function returned without error
	if errorCode != 0 {
		err = fmt.Errorf("Signature Failed!")
		return
	}

	signedMessage.Message = message
	// copies the last SignatureSize bytes of signedMessageBytes into signedMessage.Signature
	copy(signedMessage.Signature[:], signedMessageBytes[:len(signedMessage.Signature)])

	return
}

// takes as input a public key and a signed message
// returns whether the signature is valid or not
func Verify(verificationKey PublicKey, signedMessage SignedMessage) (verified bool, err error) {
	// points to unsigned message after verifying
	// the +1 is a lazy way to prevent runtime panic if message is empty
	messageBytes := make([]byte, len(signedMessage.Message)+1)
	messagePointer := (*C.uchar)(unsafe.Pointer(&messageBytes[0]))

	// points to an int so the C function can return the message length after verifying
	var messageLen uint64
	lenPointer := (*C.ulonglong)(unsafe.Pointer(&messageLen))

	// points to the signed message as input for verification
	signedMessageBytes := []byte(signedMessage.CombinedMessage())
	signedMessagePointer := (*C.uchar)(unsafe.Pointer(&signedMessageBytes[0]))

	// length of signed message, but as a C object
	signedMessageLen := C.ulonglong(len(signedMessageBytes))

	// pointer to the public key used to sign the message
	verificationKeyPointer := (*C.uchar)(unsafe.Pointer(&verificationKey[0]))

	// verify signature
	success := C.crypto_sign_open(messagePointer, lenPointer, signedMessagePointer, signedMessageLen, verificationKeyPointer)

	if success == 0 {
		verified = true
		return
	}

	verified = false
	return
}
