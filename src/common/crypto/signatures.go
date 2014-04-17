package crypto

// #cgo LDFLAGS: -lsodium
// #include <sodium.h>
import "C"

import (
	"fmt"
	"unsafe"
)

type SignedMessage struct {
	Signature Signature
	Message   string
}

// Libsodium puts the signature first and the message second, therefore so do we
func (sm *SignedMessage) CombinedMessage() (combinedMessage string) {
	combinedMessage = string(append(sm.Signature[:], []byte(sm.Message[:])...))
	return
}

// CreateKeyPair needs no input, produces a public key and secret key as output
func CreateKeyPair() (publicKey PublicKey, secretKey SecretKey, err error) {
	errorCode := C.crypto_sign_keypair((*C.uchar)(unsafe.Pointer(&publicKey[0])), (*C.uchar)(unsafe.Pointer(&secretKey[0])))
	if errorCode != 0 {
		err = fmt.Errorf("Key Creation Failed!")
		return
	}

	return
}

// Sign takes a secret key and a message, and use the secret key to sign the message.
// Sign returns a single SignedMessage struct containing a Message and a Signature
func Sign(secretKey SecretKey, message string) (signedMessage SignedMessage, err error) {
	// Points to a signed message of format signature + message after sigining
	signedMessageBytes := make([]byte, len(message)+SignatureSize)
	signedMessagePointer := (*C.uchar)(unsafe.Pointer(&signedMessageBytes[0]))

	// Points to the length of the signed message after signing
	var signatureLen uint64
	lenPointer := (*C.ulonglong)(unsafe.Pointer(&signatureLen))

	// Points to the message
	var messagePointer *C.uchar
	// Can't point to a slice of length 0
	if len(message) == 0 {
		messagePointer = (*C.uchar)(nil)
	} else {
		messageBytes := []byte(message)
		messagePointer = (*C.uchar)(unsafe.Pointer(&messageBytes[0]))
	}

	// How long the message is
	messageLen := C.ulonglong(len(message))

	// Pointer to the signature
	signaturePointer := (*C.uchar)(unsafe.Pointer(&secretKey[0]))

	// sign message
	errorCode := C.crypto_sign(signedMessagePointer, lenPointer, messagePointer, messageLen, signaturePointer)
	if errorCode != 0 {
		err = fmt.Errorf("Signature Failed!")
		return
	}

	signedMessage.Message = message
	// copy the signature from the byte slice to the Signature field of signedMessage
	copy(signedMessage.Signature[:], signedMessageBytes[:len(signedMessage.Signature)])

	return
}

// takes as input a public key and a signed message
// returns whether the signature is valid or not
func Verify(verificationKey PublicKey, signedMessage SignedMessage) (verified bool, err error) {
	// points to unsigned message after verifying
	var messagePointer *C.uchar
	messageBytes := make([]byte, len(signedMessage.Message)+1)
	if len(signedMessage.Message) == 0 {
		// must point somewhere valid, but the data won't be altered
		// can't point to [0] because the slice is empty
		messagePointer = (*C.uchar)(unsafe.Pointer(&messageBytes))
	} else {
		messagePointer = (*C.uchar)(unsafe.Pointer(&messageBytes[0]))
	}

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
	verified = success == 0
	return
}

// Only used for testing, but testing functions in multiple packages use it
func CheckKeys(publicKey PublicKey, secretKey SecretKey) (err error) {
	// create an arbitrary message
	randomMessage, err := RandomByteSlice(20)
	if err != nil {
		return
	}

	// sign the message
	signedMessage, err := Sign(secretKey, string(randomMessage))
	if err != nil {
		return
	}

	// verify the message
	verification, err := Verify(publicKey, signedMessage)
	if err != nil {
		return
	}

	// check that verification succeeded
	if !verification {
		err = fmt.Errorf("Unable to verify key!")
		return
	}

	// create an imposter verification message
	var imposterMessage SignedMessage
	imposterMessage.Message = "a"
	for i := range imposterMessage.Signature {
		imposterMessage.Signature[i] = byte(8)
	}

	// try to verify imposter message
	verification, err = Verify(publicKey, imposterMessage)
	if err != nil {
		return
	} else if verification {
		err = fmt.Errorf("Able to verify a fake message!")
		return
	}

	// test signing an empty message
	emptyMessage, err := Sign(secretKey, "")
	if err != nil {
		return
	}

	// test verifiying an empty message
	verification, err = Verify(publicKey, emptyMessage)
	if err != nil {
		return
	} else if !verification {
		err = fmt.Errorf("Could not verify empty message")
		return
	}

	// test verifying an empty message with fake signature
	emptyMessage.Signature[0] = 0
	emptyMessage.Signature[1] = 0
	verification, err = Verify(publicKey, emptyMessage)
	// unfinished

	return
}
