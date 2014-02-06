package swarm

import (
	"crypto/rand"
)

func Entropy() (randomBytes []byte, err error) {
	randomBytes = make([]byte, 256)
	_, err = rand.Read(randomBytes)
	return randomBytes, nil
}
