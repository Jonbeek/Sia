package crypto

// #include "crypto_test.c"
import "C"

// cgo can't be used in a test file, so it was necessary to
// do the cgo-required tests in a test_supplement file

// Verify PublicKeySize
func VerifyPublicKeySize() bool {
	confirmation := C.testPublicKeySize(C.int(PublicKeySize))
	return confirmation != 0
}

// Verify SecretKeySize
func VerifySecretKeySize() bool {
	confirmation := C.testSecretKeySize(C.int(SecretKeySize))
	return confirmation != 0
}

// Verify SignatureSize
func VerifySignatureSize() (verification bool) {
	confirmation := C.testSignatureSize(C.int(SignatureSize))
	return confirmation != 0
}

// Verify HashSize
func VerifyHashSize() (verification bool) {
	confirmation := C.testHashSize(C.int(HashSize))
	return confirmation != 0
}
