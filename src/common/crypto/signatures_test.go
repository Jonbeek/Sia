package crypto

import (
	"crypto/rand"
	"testing"
)

// Basic testing of key creation, signing, and verification
// Also checks that verification fails properly
func TestSigning(t *testing.T) {
	// Create a keypair
	publicKey, secretKey, err := CreateKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	// create an arbitrary message
	randomMessage := make([]byte, 1024)
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
	imposterMessage := make([]byte, 1024+SignatureSize)
	rand.Read(imposterMessage)

	// try to verify imposter message
	verification, err = Verify(publicKey, string(imposterMessage))
	if verification {
		t.Fatal("Imposter message was verified as legitimate!")
	}
}

// There should probably be some benchmarking
