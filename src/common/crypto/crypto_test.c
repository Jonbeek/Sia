#include <sodium.h>

static int testPublicKeySize(int publicKeySize) {
	if(publicKeySize != crypto_sign_PUBLICKEYBYTES) {
		return 0;
	}

	return 1;
}

static int testSecretKeySize(int secretKeySize) {
	if(secretKeySize != crypto_sign_SECRETKEYBYTES) {
		return 0;
	}

	return 1;
}

static int testSignatureSize(int signatureSize) {
	if(signatureSize != crypto_sign_BYTES) {
		return 0;
	}

	return 1;
}

static int testHashSize(int hashSize) {
	if(hashSize != crypto_hash_BYTES) {
		return 0;
	}

	return 1;
}
