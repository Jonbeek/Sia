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
	"fmt"
	"unsafe"
)

// EncodeRing takes a Sector and encodes it as a Ring: a set of common.QuorumSize Segments that include redundancy.
// The Sector's k value indicates the number of non-redundant segments, and its b value indicates the size of each segment.
// The erasure-coding algorithm requires that the original data must be k*b in size, so it is padded here as needed.
//
// The return value is a Ring.
// The first k Segments of the Ring are the original data split up.
// The remaining Segments are newly generated redundant data.
func EncodeRing(sec *common.Sector) (ring common.Ring, err error) {
	k, b := sec.GetRedundancy(), sec.GetBytesPerSegment()

	// check for legal size of k
	if k <= 0 || k >= common.QuorumSize {
		err = fmt.Errorf("k must be greater than 0 and smaller than %v", common.QuorumSize)
		return
	}

	// check for legal size of b
	if b < common.MinSegmentSize || b > common.MaxSegmentSize {
		err = fmt.Errorf("b must be greater than %v and smaller than %v", common.MinSegmentSize, common.MaxSegmentSize)
		return
	}

	// check that b is divisible by 64
	if b%64 != 0 {
		err = fmt.Errorf("b must be divisible by 64")
		return
	}

	// pad data as needed
	padding := k*b - len(sec.Data)
	paddedData := append(sec.Data, bytes.Repeat([]byte{0x00}, padding)...)

	// call c library to encode data
	m := common.QuorumSize - k
	redundantChunk := C.encodeRedundancy(C.int(k), C.int(m), C.int(b), (*C.char)(unsafe.Pointer(&paddedData[0])))
	redundantString := C.GoStringN(redundantChunk, C.int(m*b))

	// split paddedData into ring
	for i := 0; i < k; i++ {
		ring[i].Data = string(paddedData[i*b : i*b+b])
		ring[i].Index = uint8(i)
	}

	// split redundantString into ring
	for i := k; i < common.QuorumSize; i++ {
		ring[i].Data = redundantString[(i-k)*b : ((i-k)*b)+b]
		ring[i].Index = uint8(i)
	}

	// free the memory allocated by the C file
	C.free(unsafe.Pointer(redundantChunk))

	return
}

// RebuildSector takes a Sector and a set of Segments and recovers the original data.
// The Sector's k value must be equal to the number of non-redundant segments when the file was originally built.
// Because recovery is just a bunch of matrix operations, there is no way to tell if the data has been corrupted
// or if an incorrect value of k has been chosen. This error checking must happen before calling RebuildSector.
// Each Segment's Data must have the correct Index from when it was encoded.
//
// The original data is stored in the input Sector's Data field.
func RebuildSector(sec *common.Sector, segs []common.Segment) error {
	k, b := sec.GetRedundancy(), sec.GetBytesPerSegment()

	// check for legal size of k
	m := common.QuorumSize - k
	if k > common.QuorumSize || k < 1 {
		return fmt.Errorf("k must be greater than 0 but smaller than %v", common.QuorumSize)
	}

	// check for legal size of b
	if b < common.MinSegmentSize || b > common.MaxSegmentSize {
		return fmt.Errorf("b must be greater than %v and smaller than %v", common.MinSegmentSize, common.MaxSegmentSize)
	}

	// check for correct number of segments
	if len(segs) != k {
		return fmt.Errorf("wrong number of segments: expected %v, got %v", k, len(segs))
	}

	// move all data into a single slice for C
	var segmentData []byte
	var segmentIndices []uint8
	for i := range segs {
		byteSlice := []byte(segs[i].Data)

		// verify that each string is the correct length
		if len(byteSlice) != b {
			return fmt.Errorf("at least 1 Segment's Data field is the wrong length")
		}

		segmentData = append(segmentData, byteSlice...)
		segmentIndices = append(segmentIndices, segs[i].Index)

	}
	// call the recovery option
	C.recoverData(C.int(k), C.int(m), C.int(b), (*C.uchar)(unsafe.Pointer(&segmentData[0])), (*C.uchar)(unsafe.Pointer(&segmentIndices[0])))

	// remove padding introduced by EncodeRing()
	sec.Data = segmentData[:sec.GetLength()]
	return nil
}
