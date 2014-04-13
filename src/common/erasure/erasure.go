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
	"common"
	"fmt"
	"unsafe"
)

// EncodeRing takes data and produces a set of common.QuorumSize pieces that include redundancy.
// 'k' indiciates the number of non-redundant segments, and 'bytesPerSegment' indicates the size of each segment.
// 'originalData' must be 'k' * 'bytesPerSegment' in size, and should be padded before calling 'EncodeRing'.
//
// The return value is a set of strings common.QuorumSize in length.
// Each string is bytesPerSegment large.
// The first 'k' strings are the original data split up.
// The remaining strings are newly generated redundant data.
func EncodeRing(k int, bytesPerSegment int, originalData []byte) (segmentdData []string, err error) {
	// check that 'k' is legal
	if k <= 0 || k >= common.QuorumSize {
		err = fmt.Errorf("k must be greater than 0 and smaller than %v", common.QuorumSize)
		return
	}

	// check that bytesPerSegment is not too big or small
	if bytesPerSegment < common.MinSliceSize || bytesPerSegment > common.MaxSliceSize {
		err = fmt.Errorf("bytesPerSegment must be greater than %v and smaller than %v", common.MinSliceSize, common.MaxSliceSize)
		return
	}

	// check that bytesPerSegment is divisible by 8
	if bytesPerSegment%8 != 0 {
		err = fmt.Errorf("bytesPerSegment must be divisible by 8")
		return
	}

	// check that originalData is the correct size
	if len(originalData) != bytesPerSegment*k {
		err = fmt.Errorf("originalData incorrectly padded, must be of size 'bytesPerSegment' * %v - 'm'", common.QuorumSize)
		return
	}

	// call c library to encode data
	m := common.QuorumSize - k
	redundantChunk := C.encodeRedundancy(C.int(k), C.int(m), C.int(bytesPerSegment), (*C.char)(unsafe.Pointer(&originalData[0])))
	redundantString := C.GoStringN(redundantChunk, C.int(m*bytesPerSegment))

	segmentdData = make([]string, common.QuorumSize)

	// split originalData into segmentdData
	for i := 0; i < k; i++ {
		segmentdData[i] = string(originalData[i*bytesPerSegment : i*bytesPerSegment+bytesPerSegment])
	}

	// split redundantString into segmentdData
	for i := k; i < common.QuorumSize; i++ {
		segmentdData[i] = redundantString[(i-k)*bytesPerSegment : ((i-k)*bytesPerSegment)+bytesPerSegment]
	}

	// free the memory allocated by the C file
	C.free(unsafe.Pointer(redundantChunk))

	return
}

// RebuildSector takes a set of 'k' strings, each 'bytesPerSegment' in size, and recovers the original data.
// 'k' must be equal to the number of non-redundant segments when the file was originally built.
// Because recovery is just a bunch of matrix operations, there is no way to tell if the data has been corrupted
// or if an incorrect value of 'k' has been chosen. This error checking must happen before calling RebuildSector.
// The set of 'untaintedSegments' will have corresponding indicies from when they were encoded.
// There is no way to tell what the indicies are, so they must be supplied in the 'segmentIndicies' slice.
// This must be a uint8 because the C library uses a char.
//
// The output is a single byteslice that is equivalent to the data used when initially calling EncodeRing()
func RebuildSector(k int, bytesPerSegment int, untaintedSegments []string, segmentIndicies []uint8) (originalData []byte, err error) {
	// check for legal size of k and m
	m := common.QuorumSize - k
	if k > common.QuorumSize || k < 1 {
		err = fmt.Errorf("k must be greater than 0 but smaller than %v", common.QuorumSize)
		return
	}

	// check for legal size of bytesPerSegment
	if bytesPerSegment < common.MinSliceSize || bytesPerSegment > common.MaxSliceSize {
		err = fmt.Errorf("bytesPerSegment must be greater than %v and smaller than %v", common.MinSliceSize, common.MaxSliceSize)
		return
	}

	// check that input data is correct number of slices.
	if len(untaintedSegments) != k {
		err = fmt.Errorf("there must be k elements in untaintedSegments")
		return
	}

	// check that input indicies are correct number of indicies
	if len(segmentIndicies) != k {
		err = fmt.Errorf("there must be k elements in segmentIndicies")
		return
	}

	// move all data into a single slice for C
	originalData = make([]byte, 0, k*bytesPerSegment)
	for _, segment := range untaintedSegments {
		byteSlice := []byte(segment)

		// verify that each string is the correct length
		if len(byteSlice) != bytesPerSegment {
			err = fmt.Errorf("at least 1 of 'untaintedSegments' is the wrong length")
			return
		}

		originalData = append(originalData, byteSlice...)
	}

	// call the recovery option
	C.recoverData(C.int(k), C.int(m), C.int(bytesPerSegment), (*C.uchar)(unsafe.Pointer(&originalData[0])), (*C.uchar)(unsafe.Pointer(&segmentIndicies[0])))

	return
}
