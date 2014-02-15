package common

import (
	"crypto/rand"
	"hash"
	"math/big"
)

func HashedRandomData(h hash.Hash, length uint64) (hash string, data string) {
	b := make([]byte, length)
	err, _ := rand.Read(b)
	if err != nil {
		panic(err)
	}

	h.Reset()
	h.Write(b)

	hash = string(h.Sum(nil))
	data = string(b)

	return
}

func RendezvousHash(h hash.Hash, items []string, key string) (r string) {
	r = items[0]
	v := big.NewInt(0)

	t := big.NewInt(0)
	for _, i := range items {
		//Hash
		h.Reset()
		h.Write([]byte(i))

		//Convert to Number and Compare
		t.SetBytes(h.Sum(nil))
		if t.Cmp(v) > 0 {
			v.Set(t)
			r = i
		}
	}
	return
}
