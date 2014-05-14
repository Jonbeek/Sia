// Sia uses Reed-Solomon coding for error correction. This package has no
// method for error detection however, so error detection must be performed
// elsewhere.
//
// We use the repository 'Siacoin/longhair' to handle the erasure coding.
// As far as I'm aware, it's the fastest library that's open source.
// It is a fork of 'catid/longhair', and we inted to merge all changes from
// the original.
//
// Longhair is a c++ library. Here, it is cast to a C library and then called
// using cgo.
package erasure

// #include "erasure.c"
import "C"

import (
	"bytes"
	"common"
	"common/crypto"
	"fmt"
	"unsafe"
)

// EncodeRing takes a Sector and encodes it as a Ring: a set of common.QuorumSize Segments that include redundancy.
// k is the number of non-redundant segments, and b is the size of each segment. b is calculated from k.
// The erasure-coding algorithm requires that the original data must be k*b in size, so it is padded here as needed.
//
// The return value is a Ring.
// The first k Segments of the Ring are the original data split up.
// The remaining Segments are newly generated redundant data.
func EncodeRing(sec *common.Sector, k int) (ring *common.Ring, segs [common.QuorumSize]common.Segment, err error) {
	// check for legal size of k
	if k <= 0 || k >= common.QuorumSize {
		err = fmt.Errorf("k must be greater than 0 and smaller than %v", common.QuorumSize)
		return
	}

	// calculate b
	b := len(sec.Data) / k
	if b%64 != 0 {
		b += 64 - (b % 64)
	}

	// check for legal size of b
	if b < common.MinSegmentSize || b > common.MaxSegmentSize {
		err = fmt.Errorf("b must be greater than %v and smaller than %v", common.MinSegmentSize, common.MaxSegmentSize)
		return
	}

	// store encoding parameters in ring
	ring = common.NewRing(k, b, len(sec.Data))

	// pad data as needed
	padding := k*b - len(sec.Data)
	paddedData := append(sec.Data, bytes.Repeat([]byte{0x00}, padding)...)

	// call the encoding function
	m := common.QuorumSize - k
	redundantChunk := C.encodeRedundancy(C.int(k), C.int(m), C.int(b), (*C.char)(unsafe.Pointer(&paddedData[0])))
	redundantBytes := C.GoBytes(unsafe.Pointer(redundantChunk), C.int(m*b))

	// split paddedData into ring
	for i := 0; i < k; i++ {
		segs[i] = common.Segment{
			paddedData[i*b : (i+1)*b],
			uint8(i),
		}
		ring.SegHashes[i], err = crypto.CalculateHash(paddedData[i*b : (i+1)*b])
		if err != nil {
			return
		}
	}

	// split redundantString into ring
	for i := k; i < common.QuorumSize; i++ {
		segs[i] = common.Segment{
			redundantBytes[(i-k)*b : (i-k+1)*b],
			uint8(i),
		}
		ring.SegHashes[i], err = crypto.CalculateHash(redundantBytes[(i-k)*b : (i-k+1)*b])
		if err != nil {
			return
		}
	}

	// free the memory allocated by the C file
	C.free(unsafe.Pointer(redundantChunk))

	return
}

// RebuildSector takes a Ring and returns a Sector containing the original data.
// The Ring's k value must be equal to the number of non-redundant segments when the file was originally built.
// Because recovery is just a bunch of matrix operations, there is no way to tell if the data has been corrupted
// or if an incorrect value of k has been chosen. This error checking must happen before calling RebuildSector.
// Each Segment's Data must have the correct Index from when it was encoded.
func RebuildSector(ring *common.Ring, segs []common.Segment) (sec *common.Sector, err error) {
	k, b := ring.GetRedundancy(), ring.GetBytesPerSegment()
	if k == 0 && b == 0 {
		err = fmt.Errorf("could not rebuild using uninitialized encoding parameters")
		return
	}

	// check for legal size of k
	m := common.QuorumSize - k
	if k > common.QuorumSize || k < 1 {
		err = fmt.Errorf("k must be greater than 0 but smaller than %v", common.QuorumSize)
		return
	}

	// check for legal size of b
	if b < common.MinSegmentSize || b > common.MaxSegmentSize {
		err = fmt.Errorf("b must be greater than %v and smaller than %v", common.MinSegmentSize, common.MaxSegmentSize)
		return
	}

	// check for correct number of segments
	if len(segs) < k {
		err = fmt.Errorf("insufficient segments: expected at least %v, got %v", k, len(segs))
		return
	}

	// move all data into a single slice
	var segmentData []byte
	var segmentIndices []uint8
	for i := 0; i < k; i++ {
		// verify that each segment is the correct length
		if len(segs[i].Data) != b {
			err = fmt.Errorf("at least 1 Segment's Data field is the wrong length")
			return
		}

		segmentData = append(segmentData, segs[i].Data...)
		segmentIndices = append(segmentIndices, segs[i].Index)

	}
	// call the recovery function
	C.recoverData(C.int(k), C.int(m), C.int(b), (*C.uchar)(unsafe.Pointer(&segmentData[0])), (*C.uchar)(unsafe.Pointer(&segmentIndices[0])))

	// remove padding introduced by EncodeRing()
	sec, err = common.NewSector(segmentData[:ring.GetLength()])
	return
}
