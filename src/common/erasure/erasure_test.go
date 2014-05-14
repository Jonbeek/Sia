package erasure

import (
	"common"
	"common/crypto"
	"testing"
)

// Basic test for reed-solomon coding, verifies that standard input
// will produce the correct results.
func TestCoding(t *testing.T) {
	// set encoding parameters
	k := common.QuorumSize / 2
	m := common.QuorumSize - k
	b := 1024

	// create sector data
	randomBytes, err := crypto.RandomByteSlice(b * k)
	if err != nil {
		t.Fatal(err)
	}

	// create sector
	sec, err := common.NewSector(randomBytes)
	if err != nil {
		t.Fatal(err)
	}

	// encode data into a Ring
	_, segs, err := EncodeRing(sec, k)
	if err != nil {
		t.Fatal(err)
	}

	// create Ring from subset of encoded segments
	newRing := common.NewRing(k, b, len(randomBytes))
	var newSegs []common.Segment
	for i := m; i < common.QuorumSize; i++ {
		newSegs = append(newSegs, segs[i])
	}

	// recover original data
	newSec, err := RebuildSector(newRing, newSegs)
	if err != nil {
		t.Fatal(err)
	}

	// compare to hash of data when first generated
	recoveredDataHash, err := crypto.CalculateHash(newSec.Data)
	if err != nil {
		t.Fatal(err)
	} else if recoveredDataHash != sec.Hash {
		t.Fatal("recovered data is different from original data")
	}

	// In every test, we check that the hashes equal
	// every other hash that gets created. This makes
	// me uneasy.
}

// At some point, there should be a long test that explores all of the edge cases.

// There should be a fuzzing test that explores random inputs. In particular, I would
// like to fuzz the 'RebuildSector' function

// There should also be a benchmarking test here.
