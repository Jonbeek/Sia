package swarm

import (
	"testing"
)

func Test_EntropyBytes(t *testing.T) {
	_, err := EntropyBytes()
	if err != nil {
		t.Fatal(err)
	}
}

func Test_DRNGSeed(t *testing.T) {
	b := new(Block)
	b.EntropyStage2 = make(map[string][]byte)

	first, err := EntropyBytes()
	if err != nil {
		t.Fatal(err)
	}
	b.EntropyStage2["1"] = first

	second, err := EntropyBytes()
	if err != nil {
		t.Fatal(err)
	}
	b.EntropyStage2["2"] = second

	third, err := EntropyBytes()
	if err != nil {
		t.Fatal(err)
	}
	b.EntropyStage2["3"] = third

	_, err = b.DRNGSeed()
	if err != nil {
		t.Fatal(err)
	}
}
