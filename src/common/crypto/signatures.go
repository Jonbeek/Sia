package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
)

type SignedMessage struct {
	Signature Signature
	Message   string
}

// Return a string containing both the message and the prepended signature
func (sm *SignedMessage) CombinedMessage() (combinedMessage string) {
	signature := append([]byte(sm.Signature.R.String()), []byte(sm.Signature.S.String())...)
	combinedMessage = string(append(signature, []byte(sm.Message[:])...))
	return
}

// CreateKeyPair needs no input, produces a public key and secret key as output
func CreateKeyPair() (publicKey PublicKey, secretKey SecretKey, err error) {
	priv, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	secretKey = SecretKey(*priv)
	publicKey = PublicKey(priv.PublicKey)
	return
}

// Sign takes a secret key and a message, and use the secret key to sign the message.
// Sign returns a single SignedMessage struct containing a Message and a Signature
func Sign(secretKey SecretKey, message string) (signedMessage SignedMessage, err error) {
	ecdsaKey := ecdsa.PrivateKey(secretKey)
	r, s, err := ecdsa.Sign(rand.Reader, &ecdsaKey, []byte(message))
	signedMessage.Signature.R = r
	signedMessage.Signature.S = s
	signedMessage.Message = message
	return
}

// takes as input a public key and a signed message
// returns whether the signature is valid or not
func Verify(verificationKey PublicKey, signedMessage SignedMessage) (verified bool, err error) {
	ecdsaKey := ecdsa.PublicKey(verificationKey)
	verified = ecdsa.Verify(&ecdsaKey, []byte(signedMessage.Message), signedMessage.Signature.R, signedMessage.Signature.S)
	return
}

// Only used for testing, but testing functions in multiple packages use it
func CheckKeys(publicKey PublicKey, secretKey SecretKey) (err error) {
	// create an arbitrary message
	/* randomMessage, err := RandomByteSlice(20)
	if err != nil {
		return
	}

	// sign the message
	signedMessage, err := Sign(secretKey, string(randomMessage))
	if err != nil {
		return
	}

	// verify the message
	verification, err := Verify(publicKey, signedMessage)
	if err != nil {
		return
	}

	// check that verification succeeded
	if !verification {
		err = fmt.Errorf("Unable to verify key!")
		return
	}

	// create an imposter verification message
	var imposterMessage SignedMessage
	imposterMessage.Message = "a"
	for i := range imposterMessage.Signature {
		imposterMessage.Signature[i] = byte(8)
	}

	// try to verify imposter message
	verification, err = Verify(publicKey, imposterMessage)
	if err != nil {
		return
	} else if verification {
		err = fmt.Errorf("Able to verify a fake message!")
		return
	}

	// test signing an empty message
	emptyMessage, err := Sign(secretKey, "")
	if err != nil {
		return
	}

	// test verifiying an empty message
	verification, err = Verify(publicKey, emptyMessage)
	if err != nil {
		return
	} else if !verification {
		err = fmt.Errorf("Could not verify empty message")
		return
	}

	// test verifying an empty message with fake signature
	emptyMessage.Signature[0] = 0
	emptyMessage.Signature[1] = 0
	verification, err = Verify(publicKey, emptyMessage)
	// unfinished */

	return
}
