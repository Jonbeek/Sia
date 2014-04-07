package crypto

import (
	"testing"
)

// A basic test, checks for 3 constants
func TestConstants(t *testing.T) {
	if TestPublicKeySize() != true {
		t.Fatal("PublicKeySize does not match libsodium crypto_sign_PUBLICKEYBYTES")
	}

	if TestSecretKeySize() != true {
		t.Fatal("SecretKeySize does not match libsodium crypto_sign_SECRETKEYBYTES")
	}

	if TestSignatureSize() != true {
		t.Fatal("SignatureSize does not match libsodium crypto_sign_BYTES")
	}
}
