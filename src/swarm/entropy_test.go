package swarm

import (
	"bytes"
	"testing"
)

// Not sure if this should be 3 functions or 1, but if it were broken into
// 3 function, this would still be what the 3rd looks like. Basically, in order
// to test SiaRandomNumber, you have to run all the other stuff too.
func Test_EntropyFuncs(t *testing.T) {
	block := new(Block)
	block.EntropyStage2 = make(map[string][]byte)

	// Test EntropyBytes()
	first, err := EntropyBytes()
	if err != nil {
		t.Fatal(err)
	}
	block.EntropyStage2["1"] = first

	second, err := EntropyBytes()
	if err != nil {
		t.Fatal(err)
	}
	block.EntropyStage2["2"] = second

	third, err := EntropyBytes()
	if err != nil {
		t.Fatal(err)
	}
	block.EntropyStage2["3"] = third

	// Test DRNGSeed()
	seed, err := block.DRNGSeed()
	if err != nil {
		t.Fatal(err)
	}

	// Test SiaRandomNumber()
	blockChain := new(BlockChain)
	blockChain.DRNGSeed = make([]byte, EntropyVolume)
	copied := copy(blockChain.DRNGSeed, seed)
	siaRand, err := blockChain.SiaRandomNumber()
	if err != nil {
		t.Fatal(err)
	}

	// Check that the correct volume of data was copied over
	if copied != EntropyVolume {
		t.Fatal("SiaRandomNumber produces entropy of the wrong size!")
	}

	// Check that the seed was hashed to produce the random number
	if bytes.Compare(siaRand, seed) == 0 {
		t.Fatal("DRNGSeed was not hashed when producing siaRand!")
	}

	// Check that DRNGSeed was updated
	if bytes.Compare(siaRand, blockChain.DRNGSeed) != 0 {
		t.Fatal("DRNGSeed was not updated with producing siaRand!")
	}
}
