package swarm

import (
	"bytes"
	"testing"
)

func TestEntropyGeneration(t *testing.T) {
	entropy, err := EntropyGeneration()
	if err != nil {
		t.Fatal(err)
	}

	if len(entropy) != EntropyVolume {
		t.Fatal("EntropyGeneration generated an incorrect volume of entropy volume. Expected %v, got %v", EntropyVolume, len(entropy))
	}

	t.Log("EntropyGeneration:", entropy)

	// Future: test multiple generations and run entropy scans on them
}

func TestDRNGSeed(t *testing.T) {
	var hb1 Heartbeat
	var hb2 Heartbeat
	var hb3 Heartbeat

	entropy1, err1 := EntropyGeneration()
	entropy2, err2 := EntropyGeneration()
	entropy3, err3 := EntropyGeneration()

	if err1 != nil {
		t.Fatal(err1)
	}

	if err2 != nil {
		t.Fatal(err2)
	}

	if err3 != nil {
		t.Fatal(err3)
	}

	hb1.EntropyStage2 = entropy1
	hb2.EntropyStage2 = entropy2
	hb3.EntropyStage2 = entropy3

	hbSlice := make([]Heartbeat, 3)
	hbSlice[0] = hb1
	hbSlice[1] = hb2
	hbSlice[2] = hb3

	seed, err := DRNGSeed(hbSlice)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("DRNGSeed:", seed)

	// Future: do this a bunch of times and run entropy tests on the seeds
}

func TestSiaRandomNumber(t *testing.T) {
	var hb Heartbeat
	entropy, err := EntropyGeneration()
	if err != nil {
		t.Fatal(err)
	}

	hb.EntropyStage2 = entropy
	hbSlice := make([]Heartbeat, 1)
	hbSlice[0] = hb

	seed, err := DRNGSeed(hbSlice)
	if err != nil {
		t.Fatal(err)
	}

	var s State
	s.DRNGSeed = make([]byte, EntropyVolume)
	copy(s.DRNGSeed, seed)
	rand, err := s.SiaRandomNumber() // SiaRandomNumber() currently will never produce an error that is not nil
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(rand, seed) == 0 {
		t.Fatal("SiaRandomNumber did not hash the random seed")
	}

	if bytes.Compare(rand, s.DRNGSeed) != 0 {
		t.Fatal("SiaRandomNumber did not update DRNGSeed")
	}

	// in the future, may want tests that verify these functions follow the protocol exactly
	// meaning the hashed values match an expected specific value
}
