package swarm

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
)

/* The blockchain requires that hosts generate entropy for each heartbeat,
this funciton will produce a byte slice of the correct size filled with
psuedo-random data. Entropy() generates the entropy */

func Entropy() (randomBytes []byte, err error) {
	randomBytes = make([]byte, EntropyVolume)
	_, err = rand.Read(randomBytes)
	return
}

/* Each block, all of the entropy introduced in the heartbeats must be merged
to produce a single seed for the random number generator. MergeEntropy fulfills
this requirement according to the protocol specification */

/* MergeEntropy assumes that the set of entropy strings in the block have
already been sorted according to the protocol specification */

func (b Block) MergeEntropy() (seed []byte, err error) {
	index := 0
	base := make([]byte, len(b.EntropyStage2)*EntropyVolume)
	for _, value := range b.EntropyStage2 {
		if len(value) != EntropyVolume {
			err = errors.New("EntropyStage2 contains invalid values!")
			return nil, err
		}

		for i := 0; i < len(value); i++ {
			base[index] = value[i]
			index++
		}
	}

	hash := sha256.New()
	hash.Write(base)
	return hash.Sum(nil), nil
}
