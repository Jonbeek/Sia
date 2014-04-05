package erasure

import (
	"bytes"
	"common"
	"crypto/rand"
	"crypto/sha256"
	"testing"
)

// Basic test for reed-solomon coding, verifies that standard input
// will produce the correct results.
func TestCoding(t *testing.T) {
	k := 100
	m := common.QUORUMSIZE - k
	bytesPerSegment := 1024

	// generate a random original file
	numRandomBytes := bytesPerSegment * k
	randomBytes := make([]byte, numRandomBytes)
	rand.Read(randomBytes)

	// get hash of original file
	randomBytesHash := common.Hash(sha256.New(), string(randomBytes))

	// encode original file into a data ring
	ringSegments, err := EncodeRing(k, bytesPerSegment, randomBytes)
	if err != nil {
		t.Fatal(err)
	}

	// verify that first k segments are still original data
	originalDataHash := common.Hash(sha256.New(), string(randomBytes))
	if !(bytes.Equal([]byte(originalDataHash), []byte(randomBytesHash))) {
		t.Fatal("original data was modified after caling EncodeRing!")
	}

	// reduce file to a set of k segments and print those segments out
	remainingSegments := make([]string, k)
	segmentIndicies := make([]uint8, k)
	for i := m; i < common.QUORUMSIZE; i++ {
		remainingSegments[i-m] = ringSegments[i]
		segmentIndicies[i-m] = uint8(i)
	}

	// recover original data
	recoveredData, err := RebuildSector(k, bytesPerSegment, remainingSegments, segmentIndicies)
	if err != nil {
		t.Fatal(err)
	}

	// compare to hash of data when first generated
	recoveredDataHash := common.Hash(sha256.New(), string(recoveredData))
	if !(bytes.Equal([]byte(recoveredDataHash), []byte(randomBytesHash))) {
		t.Fatal("recovered data is different from original data")
	}
}

// At some point, there should be a long test that explores all of the edge cases.

// There should be a fuzzing test that explores random inputs. In particular, I would
// like to fuzz the 'RebuildSector' function

// There should also be a benchmarking test here.
