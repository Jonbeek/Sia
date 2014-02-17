package common

import (
	"crypto/sha256"
	"testing"
)

func TestRendezvousHash(t *testing.T) {
	h := "test"
	c := make([]string, 0)
	c = append(c, "foo")
	r := RendezvousHash(sha256.New(), c, h)

	if r != "foo" {
		t.Log(r)
		t.Log(h)
		t.Fatal("Got back a item we didn't give to RendezvousHash")
	}
}
