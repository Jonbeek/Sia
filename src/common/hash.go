package common

import (
	"hash"
	"math/big"
)

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
