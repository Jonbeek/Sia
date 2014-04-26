package common

import (
	"common/crypto"
	"testing"
)

// Marshal an identifier, and see if it unmarshalls to the same value
// Check that bad input is handled correctly
// Fuzz over all possible values of Identifier during long tests
func TestIdentifierMarshalling(t *testing.T) {
	// marshal, unmarshal, verify at the zero value
	var id Identifier
	mid := id.Marshal()
	uid, err := UnmarshalIdentifier(mid)
	if err != nil {
		t.Fatal(err)
	}
	if uid != id {
		t.Fatal("Identifier not equal after marshalling and unmarshalling")
	}

	// test unmarshall for bad input
	var bad []byte
	uid, err = UnmarshalIdentifier(bad)
	if err == nil {
		t.Error("Accepted an empty slice as input!")
	}
	bad = make([]byte, 2)
	uid, err = UnmarshalIdentifier(bad)
	if err == nil {
		t.Error("UnmarshalIdentifier accepted a slice of length 2!")
	}

	// test all possible values of identifier
	if testing.Short() {
		t.Skip()
	}

	for i := byte(0); i < byte(255); i++ {
		id = Identifier(i)
		mid = id.Marshal()
		uid, err := UnmarshalIdentifier(mid)
		if err != nil {
			t.Error("Fuzzing Error: id =", id)
			t.Error(err)
			continue
		}
		if uid != id {
			t.Error("Fuzzing Error: uid does not equal id for id =", id)
		}
	}
}

// Marshal an Address and see if it unmarshalles to the same value
// Check that bad input is handled correctly
// fuzz over a bunch of random values for Address during long tests
func TestAddressMarshalling(t *testing.T) {
	// marshal, unmarshal, verify at the zero value
	var a Address
	ma := a.Marshal()
	ua, err := UnmarshalAddress(ma)
	if err != nil {
		t.Fatal(err)
	}
	if ua != a {
		t.Fatal("Address not equal after marshalling and unmarshalling!")
	}

	// test that Address marshal handles bad inputs
	var bad []byte
	ua, err = UnmarshalAddress(bad)
	if err == nil {
		t.Error("UnmarshalAddress accepted an unitialized []byte")
	}
	bad = make([]byte, 4)
	ua, err = UnmarshalAddress(bad)
	if err == nil {
		t.Error("UnmarshalAddress accepted a []byte too short to be an address")
	}

	// test marshalling a bunch of random Addresses
	if testing.Short() {
		t.Skip()
	}
	for i := 0; i < 10000; i++ {
		randomBytes, err := crypto.RandomByteSlice(crypto.RandomInt(100) + 1)
		a.Id = Identifier(randomBytes[0]) // random id
		a.Host = string(randomBytes[1:])  // random host
		a.Port = crypto.RandomInt(65536)  // random port

		ma = a.Marshal()
		ua, err = UnmarshalAddress(ma)
		if err != nil {
			t.Error(a.Id, " ", a.Host, " ", a.Port)
			t.Error(err)
		}
		if ua != a {
			t.Error("Fuzzing Test Failed: Marshalled and Unmarshalled address not identical")
			t.Error(a.Id, " ", a.Host, " ", a.Port)
		}
	}
}
