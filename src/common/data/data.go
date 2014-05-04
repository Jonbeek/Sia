package data

import (
	"bytes"
	"common"
	"common/crypto"
	"common/erasure"
)

// A Segment is an erasure-coded piece of a Ring, containing both the data and its corresponding index.
type Segment struct {
	Data  string
	Index uint8
}

// A Sector is a block of data, along with its erasure-coding parameters.
// K is the number of redundant segments, and B is the number of bytes per segment.
type Sector struct {
	Data   []byte
	Hash   crypto.Hash
	length int
	k, b   int
}

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

// Encode encodes a Sector using its erasure-coding parameters, returning the erasure-coded segments.
// This may require padding the Sector data.
func (s *Sector) Encode() (segs [common.QuorumSize]Segment, err error) {
	padding := s.k*s.b - len(s.Data)
	s.Data = append(s.Data, bytes.Repeat([]byte{0x00}, padding)...)
	data, err := erasure.EncodeRing(s.k, s.b, s.Data)
	if err != nil {
		return
	}
	for i := range segs {
		segs[i].Data = data[i]
		segs[i].Index = uint8(i)
	}
	return
}

// Rebuild restores the original data of a sector from its erasure-coded segments.
// It trims the rebuilt data to remove padding introduced by Encode()
// TODO: make this take a []Segment, and rewrite RebuildSector
func (s *Sector) Rebuild(segments []string, indices []uint8) (err error) {
	data, err := erasure.RebuildSector(s.k, s.b, segments[:s.k], indices[:s.k])
	if err != nil {
		return
	}
	s.Data = data[:s.length]
	return
}
