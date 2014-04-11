package crypto

import (
	"testing"
)

// test each type of hash using string "foo" and compare to a
// reference value
func TestHashing(t *testing.T) {
	// in bytes, the hash of string "foo"
	referenceHash := Hash{247, 251, 186, 110, 6, 54, 248, 144, 229, 111, 187, 243, 40, 62, 82, 76, 111, 163, 32, 74, 226, 152, 56, 45, 98, 71, 65, 208, 220, 102, 56, 50, 110, 40, 44, 65, 190, 94, 66, 84, 216, 130, 7, 114, 197, 81, 138, 44, 90, 140, 12, 127, 126, 218, 25, 89, 74, 126, 181, 57, 69, 62, 30, 215}

	// in bytes, the truncated hash of string "foo"
	referenceTruncatedHash := TruncatedHash{247, 251, 186, 110, 6, 54, 248, 144, 229, 111, 187, 243, 40, 62, 82, 76, 111, 163, 32, 74, 226, 152, 56, 45, 98, 71, 65, 208, 220, 102, 56, 50}

	// compute hash and compare to reference
	hash, err := CalculateHash([]byte("foo"))
	if err != nil {
		t.Fatal(err)
	} else if hash != referenceHash {
		t.Fatal("Hash producing unexpected value")
	}

	// compute truncated hash and compare to reference
	tHash, err := CalculateTruncatedHash([]byte("foo"))
	if err != nil {
		t.Fatal(err)
	} else if tHash != referenceTruncatedHash {
		t.Fatal("Truncated hash producing unexpected value")
	}
}
