package common

import (
	"common/crypto"
)

// A Sector is a logical block of data.
type Sector struct {
	Data []byte
	Hash crypto.Hash
}

// A Ring is an erasure-coded Sector, along with the parameters used to encode it.
// k is the number of non-redundant segments, and b is the number of bytes per segment.
type Ring struct {
	Hosts     Quorum
	SegHashes [QuorumSize]crypto.Hash
	k, b      int
	length    int
}

// A Segment is an erasure-coded piece of a Ring, containing both the data and its corresponding index.
type Segment struct {
	Data  []byte
	Index uint8
}

// NewSector creates a Sector from data.
func NewSector(data []byte) (s *Sector, err error) {
	// calculate hash
	hash, err := crypto.CalculateHash(data)

	s = &Sector{
		Data: data,
		Hash: hash,
	}
	return
}

// NewRing creates an empty Ring using the specified encoding parameters.
func NewRing(k, b, length int) *Ring {
	return &Ring{
		k:      k,
		b:      b,
		length: length,
	}
}

func (r *Ring) GetRedundancy() int {
	return r.k
}

func (r *Ring) GetBytesPerSegment() int {
	return r.b
}

func (r *Ring) GetLength() int {
	return r.length
}
