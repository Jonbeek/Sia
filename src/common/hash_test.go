package common

import (
	"crypto/sha256"
	"testing"
)

func TestRendezvousHash(t *testing.T) {
	c := []string{"foo", "bar"}
	r := RendezvousHash(sha256.New(), c, "test")

	if r != "bar" {
		t.Log(r)
		t.Log("test")
		t.Fatal("Got wrong host")
	}

	r = RendezvousHash(sha256.New(), c, "foofootest")

	if r != "foo" {
		t.Log(r)
		t.Log("foofootest")
		t.Fatal("Got wrong host")
	}
}
