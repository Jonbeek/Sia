package common

import (
	"crypto/rand"
	"hash"
	"math/big"
)

func Hash(h hash.Hash, data string) string {
	h.Reset()
	h.Write([]byte(data))
	return string(h.Sum(nil))
}

func HashedRandomData(h hash.Hash, length uint64) (hash string, data string) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	return Hash(h, string(b)), string(b)
}

func RendezvousHash(h hash.Hash, items []string, key string) (r string) {
	r = items[0]
	v := big.NewInt(0)

	t := big.NewInt(0)
	for _, i := range items {
		//Hash
		h.Reset()
		h.Write([]byte(i))
		h.Write([]byte(key))

		//Convert to Number and Compare
		t.SetBytes(h.Sum(nil))
		if t.Cmp(v) > 0 {
			v.Set(t)
			r = i
		}
	}
	return
}
