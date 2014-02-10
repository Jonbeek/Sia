package swarm

import (
	"crypto/rand"
	"crypto/sha256"
)

/* The blockchain requires that hosts generate entropy for each heartbeat,
this funciton will produce a byte slice of the correct size filled with
psuedo-random data. Entropy() generates the entropy */

func EntropyBytes() (randomBytes []byte, err error) {
	randomBytes = make([]byte, EntropyVolume)
	_, err = rand.Read(randomBytes)
	return
}

/* Each block, all of the entropy introduced in the heartbeats must be merged
to produce a single seed for the random number generator. MergeEntropy fulfills
this requirement according to the protocol specification */

/* MergeEntropy assumes that the set of entropy strings in the block have
already been sorted according to the protocol specification */

func (b Block) DRNGSeed() (seed []byte, err error) {
	var base []byte
	base = nil

	for _, value := range b.EntropyStage2 {
		base = append(base, value...)
	}

	hash := sha256.New()
	hash.Write(base)
	seed = hash.Sum(nil)
	return
}

func (b BlockChain) SiaRandomNumber() (randomNumber []byte, err error) {
	hash := sha256.New()
	hash.Write(b.DRNGSeed)
	randomNumber = hash.Sum(nil)
	copy(b.DRNGSeed, randomNumber) // might need error checking
	return
}
