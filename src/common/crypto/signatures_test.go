package crypto

import (
	"crypto/rand"
	"testing"
)

// Basic testing of key creation, signing, and verification
// Implicitly tests SignedMessage.CombinedMessage()
func TestSigning(t *testing.T) {
	// Create a keypair
	publicKey, secretKey, err := CreateKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	// create an arbitrary message
	randomMessage := make([]byte, 20)
	rand.Read(randomMessage)

	// sign the message
	signedMessage, err := Sign(secretKey, string(randomMessage))
	if err != nil {
		t.Fatal(err)
	}

	// verify the message
	verification, err := Verify(publicKey, signedMessage)
	if err != nil {
		t.Fatal(err)
	}

	// check that verification succeeded
	if !verification {
		t.Fatal("Verification of signature failed!")
	}

	// create an imposter verification message
	var imposterMessage SignedMessage
	imposterMessage.Message = "a"
	for i := range imposterMessage.Signature {
		imposterMessage.Signature[i] = byte(8)
	}

	// try to verify imposter message
	verification, err = Verify(publicKey, imposterMessage)
	if verification {
		t.Fatal("Imposter message was verified as legitimate!")
	}

	// test signing an empty message
	signedMessage, err = Sign(secretKey, "")
	if err != nil {
		t.Fatal(err)
	}

	// test verifying an empty message
	imposterMessage.Message = ""
	verification, err = Verify(publicKey, imposterMessage)
}

// There should probably be some benchmarking
