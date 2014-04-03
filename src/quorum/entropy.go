package quorum

import (
	"common"
	"crypto/rand"
	"crypto/sha256"
)

// special case of random; needs an explicit number of bytes
func EntropyGeneration() (entropy string, err error) {
	randomBytes := make([]byte, common.ENTROPYVOLUME)
	_, err = rand.Read(randomBytes)
	entropy = string(randomBytes)
	return
}

// Generates a DRNGSeed, given a presorted slice of heartbeats
func DRNGSeed(heartbeats []Heartbeat) (seed string, err error) {
	var base []byte
	base = nil

	for _, value := range heartbeats {
		base = append(base, value.EntropyStage2...)
	}

	hash := sha256.New()
	hash.Write(base)
	seed = string(hash.Sum(nil))
	return
}

func (b Blockchain) SiaRandomNumber() (randomNumber []byte, err error) {
	hash := sha256.New()
	hash.Write(b.DRNGSeed)
	randomNumber = hash.Sum(nil)
	copy(b.DRNGSeed, randomNumber) // might need error checking
	return
}
