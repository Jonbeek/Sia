package swarm

import (
	"testing"
)

func TestEntropy(t *testing.T) {
	entropy, err := Entropy()
	if err != nil {
		t.FailNow()
	}

	t.Log(entropy)
}
