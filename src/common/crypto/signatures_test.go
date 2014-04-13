package crypto

import (
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

	err = CheckKeys(publicKey, secretKey)
	if err != nil {
		t.Fatal(err)
	}
}

// There should probably be some benchmarking
