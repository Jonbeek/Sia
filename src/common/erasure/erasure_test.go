package erasure

import (
	"common"
	"common/crypto"
	"testing"
)

// Basic test for reed-solomon coding, verifies that standard input
// will produce the correct results.
func TestCoding(t *testing.T) {
	k := 100
	m := common.QuorumSize - k
	bytesPerSegment := 1024

	randomBytes, err := crypto.RandomByteSlice(bytesPerSegment * k)
	if err != nil {
		t.Fatal(err)
	}

	// get hash of original file
	randomBytesHash, err := crypto.CalculateHash(randomBytes)
	if err != nil {
		t.Fatal(err)
	}

	// encode original file into a data ring
	ringSegments, err := EncodeRing(k, bytesPerSegment, randomBytes)
	if err != nil {
		t.Fatal(err)
	}

	// verify that first k segments are still original data
	originalDataHash, err := crypto.CalculateHash(randomBytes)
	if err != nil {
		t.Fatal(err)
	} else if originalDataHash != randomBytesHash {
		t.Fatal("original data was modified after caling EncodeRing!")
	}

	// reduce file to a set of k segments and print those segments out
	remainingSegments := make([]string, k)
	segmentIndicies := make([]uint8, k)
	for i := m; i < common.QuorumSize; i++ {
		remainingSegments[i-m] = ringSegments[i]
		segmentIndicies[i-m] = uint8(i)
	}

	// recover original data
	recoveredData, err := RebuildSector(k, bytesPerSegment, remainingSegments, segmentIndicies)
	if err != nil {
		t.Fatal(err)
	}

	// compare to hash of data when first generated
	recoveredDataHash, err := crypto.CalculateHash(recoveredData)
	if err != nil {
		t.Fatal(err)
	} else if recoveredDataHash != randomBytesHash {
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
