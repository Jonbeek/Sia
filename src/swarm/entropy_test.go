package swarm

import (
	"testing"
)

func Test_Entropy(t *testing.T) {
	entropy1, err := Entropy()
	if err != nil {
		t.Fatal(err)
	}

	entropy2, err := Entropy()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(entropy1)
	t.Log(entropy2)
}

func Test_MergeEntropy(t *testing.T) {
	b := new(Block)
	b.EntropyStage2 = make(map[string][]byte)

	first, err := Entropy()
	if err != nil {
		t.Fatal(err)
	}
	b.EntropyStage2["1"] = first
	t.Log(first)

	second, err := Entropy()
	if err != nil {
		t.Fatal(err)
	}
	b.EntropyStage2["2"] = second
	t.Log(second)

	third, err := Entropy()
	if err != nil {
		t.Fatal(err)
	}
	b.EntropyStage2["3"] = third
	t.Log(third)

	seed, err := b.MergeEntropy()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(seed)
}
