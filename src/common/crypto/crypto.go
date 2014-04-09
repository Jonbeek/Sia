// package crypto uses libsodium and manages all of the crypto
// for Sia. It has an explicit typing system that uses byte
// arrays matching the sizes specified by the libsodium constants.
package crypto

const (
	// sizes in bytes
	PublicKeySize     int = 32
	SecretKeySize     int = 64
	SignatureSize     int = 64
	HashSize          int = 64
	TruncatedHashSize int = 32
)

type PublicKey [PublicKeySize]byte
type SecretKey [SecretKeySize]byte
type Signature [SignatureSize]byte
type Hash [HashSize]byte
type TruncatedHash [TruncatedHashSize]byte
