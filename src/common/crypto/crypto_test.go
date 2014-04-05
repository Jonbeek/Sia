package crypto

import (
	"testing"
)

// calls createKeyPair and checks the error
func TestCreateKeyPair(t *testing.T) {
	_, _, err := createKeyPair()
	if err != nil {
		t.Fatal(err)
	}
}
