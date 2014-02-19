package swarm

import (
	"common/crypto"
)

// special case of random; needs an explicit number of bytes
func EntropyGeneration() (randomBytes []byte, err error) {
	randomBytes, err = crypto.Random(EntropyVolume)
	return
}

// Generates a DRNGSeed, accepting a presorted slice of heartbeats as input
func DRNGSeed(blockEntropy []Heartbeat) (seed []byte, err error) {
	var base []byte
	base = nil

	for _, value := range blockEntropy {
		base = append(base, value.EntropyStage2...)
	}

	seed = crypto.Hash(base)
	return
}

// Produces a random number given a State and advances the state random number
func (s *StateSteady) SiaRandomNumber() (randomNumber []byte, err error) {
	randomNumber = crypto.Hash([]byte(s.DRNGSeed))
	s.DRNGSeed = string(randomNumber)

	return
}
