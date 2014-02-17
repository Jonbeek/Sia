package swarm

import (
	"bytes"
	"common/crypto"
	"testing"
)

func TestNewHeartbeat(t *testing.T) {
	var s StateSteady
	entropy, err := EntropyGeneration()
	if err != nil {
		t.Fatal(err)
	}

	// did this so that byte slice sizes are guaranteed to line up
	verificationEntropy := make([]byte, EntropyVolume)

	// might want to error test this copy()... but later I guess
	bytesCopied := copy(verificationEntropy, entropy)
	if bytesCopied != EntropyVolume {
		t.Fatal("Did not copy the correct number of bytes during entropy verification")
	}

	s.SecretEntropy = entropy
	hb, err := s.NewHeartbeat()
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(s.SecretEntropy, verificationEntropy) == 0 {
		t.Fatal("SecretEntropy was not altered during heartbeat creation!")
	}

	if bytes.Compare(hb.EntropyStage2, verificationEntropy) != 0 {
		t.Fatal("hb.EntropyStage2 contains the wrong value!")
	}

	entropyStage1 := crypto.Hash(s.SecretEntropy)
	if bytes.Compare(entropyStage1, hb.EntropyStage1) != 0 {
		t.Fatal("hb.EntropyStage1 does not match the hash of s.SecretEntropy")
	}

	// When testing, make sure that
	//	hb.EntropyStage1
	//	hb.EntropyStage2
	//	s.SecretEntropy
	// all adjust as desired, that they equal each other where they should and
	// that they don't equal each other where they should be different
}
