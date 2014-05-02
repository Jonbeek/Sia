package quorum

import (
	"testing"
)

func TestRandomPlacement(t *testing.T) {
	s := new(State)
	buckets, err := s.RandomPlacement(1)
	if len(buckets) == 0 {
		t.Fatal("Bucket Size = 0!")
	}
	if err != nil {
		t.Fatal("Failed RandomPlacement of 1")
	}

	buckets, err = s.RandomPlacement(0)
	if err != nil {
		t.Fatal("Failed to place 0!")
	}

	buckets, err = s.RandomPlacement(-1)
	if err == nil {
		t.Fatal("Did not produce error for negative number!")
	}

	buckets, err = s.RandomPlacement(9000)
	if err != nil {
		t.Fatal("Failed RandomPlacement of 9000")
	}
	total := 0
	for i := range buckets {
		total += buckets[i]
	}
	if total != 9000 {
		t.Fatal("Sum of buckets does not equal total given")
	}
}
