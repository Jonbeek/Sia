package erasure

import (
	"common"
	"crypto/rand"
	"crypto/sha256"
	"testing"
)

func TestCoding(t *testing.T) {
	// 1. generate a bunch of random data
	// 2. hash the data, so you know what it looks like
	// 2. call encode on the random data
	// 3. delete some of the pieces of the file, leaving just enough to rebuild it (delete random pieces)
	// 4. rebuild the file
	// 5. verify that the rebuilt file is the same as the original file

	k := 50
	m := common.SWARMSIZE - k
	bytesPerSlice := 1024

	numRandomBytes := bytesPerSlice * k
	randomBytes := make([]byte, numRandomBytes)
	rand.Read(randomBytes)

	randomBytesHash := common.Hash(sha256.New(), string(randomBytes))
	t.Log(randomBytesHash)

	_, err := EncodeRing(randomBytes, m, bytesPerSlice)
	if err != nil {
		t.Fatal(err)
	}
}
