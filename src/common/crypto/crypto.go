// package crypto uses libsodium and manages all of the crypto
// for Sia. It has an explicit typing system that uses byte
// arrays matching the sizes specified by the libsodium constants.
package crypto

import (
	"crypto/ecdsa"
	"math/big"
)

const (
	// sizes in bytes
	HashSize          int = 64
	TruncatedHashSize int = 32
)

/*type PublicKey [PublicKeySize]byte
type SecretKey [SecretKeySize]byte
type Signature [SignatureSize]byte */
type Signature struct {
	R, S *big.Int
}
type PublicKey ecdsa.PublicKey
type SecretKey ecdsa.PrivateKey
type Hash [HashSize]byte
type TruncatedHash [TruncatedHashSize]byte
