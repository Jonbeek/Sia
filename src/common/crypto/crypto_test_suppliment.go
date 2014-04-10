package crypto

// #include "crypto_test.c"
import "C"

// cgo can't be used in a test file, so it was necessary to
// do the cgo-required tests in a test_supplement file

// Verify PublicKeySize
func VerifyPublicKeySize() (verification bool) {
	confirmation := C.testPublicKeySize(C.int(PublicKeySize))

	if confirmation == 0 {
		verification = false
		return
	}

	verification = true
	return
}

// Verify SecretKeySize
func VerifySecretKeySize() (verification bool) {
	confirmation := C.testSecretKeySize(C.int(SecretKeySize))

	if confirmation == 0 {
		verification = false
		return
	}

	verification = true
	return
}

// Verify SignatureSize
func VerifySignatureSize() (verification bool) {
	confirmation := C.testSignatureSize(C.int(SignatureSize))

	if confirmation == 0 {
		verification = false
		return
	}

	verification = true
	return
}

// Verify HashSize
func VerifyHashSize() (verification bool) {
	confirmation := C.testHashSize(C.int(HashSize))

	if confirmation == 0 {
		verification = false
		return
	}

	verification = true
	return
}
