package erasure

import (
	"bytes"
	"common"
	//"common/crypto"
	"crypto/rand"   // should use rand from common/crypto
	"crypto/sha256" // should hash from common/crypto
	"encoding/hex"  // will be removed after switching to common/crypto
	"hash"          // will be removed after switching to common/crypto
	"testing"
)

// just a patch function so we can run the erasure coding tests
func Hash(h hash.Hash, data string) string {
	h.Reset()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// Basic test for reed-solomon coding, verifies that standard input
// will produce the correct results.
func TestCoding(t *testing.T) {
	k := 100
	m := common.QuorumSize - k
	bytesPerSegment := 1024

	// generate a random original file
	numRandomBytes := bytesPerSegment * k
	randomBytes := make([]byte, numRandomBytes)
	rand.Read(randomBytes)

	// get hash of original file
	randomBytesHash := Hash(sha256.New(), string(randomBytes))

	// encode original file into a data ring
	ringSegments, err := EncodeRing(k, bytesPerSegment, randomBytes)
	if err != nil {
		t.Fatal(err)
	}

	// verify that first k segments are still original data
	originalDataHash := Hash(sha256.New(), string(randomBytes))
	if !(bytes.Equal([]byte(originalDataHash), []byte(randomBytesHash))) {
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
	recoveredDataHash := Hash(sha256.New(), string(recoveredData))
	if !(bytes.Equal([]byte(recoveredDataHash), []byte(randomBytesHash))) {
		t.Fatal("recovered data is different from original data")
	}
}

// At some point, there should be a long test that explores all of the edge cases.

// There should be a fuzzing test that explores random inputs. In particular, I would
// like to fuzz the 'RebuildSector' function

// There should also be a benchmarking test here.
