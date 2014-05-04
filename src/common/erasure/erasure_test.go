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
	sec.SetRedundancy(k)

	// encode data into a Ring
	ring, err := EncodeRing(sec)
	if err != nil {
		t.Fatal(err)
	}

	// clear sector data
	sec.Data = []byte{0x00}

	// reduce data to a set of k segments
	remainingSegments := make([]common.Segment, k)
	for i := m; i < common.QuorumSize; i++ {
		remainingSegments[i-m] = ring[i]
	}

	// recover original data
	err = RebuildSector(sec, remainingSegments)
	if err != nil {
		t.Fatal(err)
	}

	// compare to hash of data when first generated
	recoveredDataHash, err := crypto.CalculateHash(sec.Data)
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
