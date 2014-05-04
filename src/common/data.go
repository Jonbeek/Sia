package common

import (
	"common/crypto"
)

// A Segment is an erasure-coded piece of a Ring, containing both the data and its corresponding index.
type Segment struct {
	Data  string
	Index uint8
}

// A Sector is a block of data, along with its erasure-coding parameters.
// k is the number of non-redundant segments, and b is the number of bytes per segment.
type Sector struct {
	Data   []byte
	Hash   crypto.Hash
	length int
	k, b   int
}

// A Ring is an array of QuorumSize Segments, ready for distribution across a Quorum.
type Ring [QuorumSize]Segment

// NewSector creates a Sector from data.
func NewSector(data []byte) (s *Sector, err error) {
	// calculate hash
	hash, err := crypto.CalculateHash(data)

	s = &Sector{
		data,
		hash,
		len(data),
		0, 0,
	}
	return
}

// SetRedundancy sets the erasure-coding parameters of a Sector based on a provided k value.
func (s *Sector) SetRedundancy(k int) {
	s.k = k
	s.b = len(s.Data) / s.k
	if s.b%64 != 0 {
		s.b += 64 - (s.b % 64)
	}
}

func (s *Sector) GetRedundancy() int {
	return s.k
}

func (s *Sector) GetBytesPerSegment() int {
	return s.b
}

func (s *Sector) GetLength() int {
	return s.length
}
