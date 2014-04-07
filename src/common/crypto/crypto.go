// Crypto is the place where all cryptographic tools
// go. Crypto tries to use exclusively code that follows
// cryptographic best practices, and is secure against
// all types of attacks.
//
// This primarily means using libsodium
package crypto

const (
	// sizes in bytes
	PublicKeySize int = 32
	SecretKeySize int = 64
	SignatureSize int = 64
)

type PublicKey string
