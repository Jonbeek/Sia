package erasure

// #cgo LDFLAGS: /home/david/git/Sia/src/common/erasure/longhair/bin/liblonghair.a -lstdc++
// #include "bridge.c"
import "C"

import (
	"common"
	"fmt"
	"unsafe"
)

func EncodeRing(originalData []byte, k int, bytesPerSlice int) (slicedData []string, err error) {
	// check that 'k' is legal
	if k <= 0 || k >= common.SWARMSIZE {
		err = fmt.Errorf("k must be greater than 0 and smaller than %v", common.SWARMSIZE)
		return
	}

	// check that bytesPerSlice is not too big or small
	if bytesPerSlice < common.MINSLICESIZE || bytesPerSlice > common.MAXSLICESIZE {
		err = fmt.Errorf("bytesPerSlice must be greater than %v and smaller than %v", common.MINSLICESIZE, common.MAXSLICESIZE)
		return
	}

	// check that bytesPerSlice is divisible by 8
	if bytesPerSlice%8 != 0 {
		err = fmt.Errorf("bytesPerSlice must be divisible by 8")
		return
	}

	// check that originalData is the correct size
	if len(originalData) != bytesPerSlice*k {
		err = fmt.Errorf("originalData incorrectly padded, must be of size 'bytesPerSlice' * %v - 'm'", common.SWARMSIZE)
		return
	}

	// call c library to encode data
	m := common.SWARMSIZE - k
	redundantChunk := C.encodeRedundancy(C.int(k), C.int(m), C.int(bytesPerSlice), (*C.char)(unsafe.Pointer(&originalData[0])))
	redundantString := C.GoStringN(redundantChunk, C.int(m*bytesPerSlice))

	slicedData = make([]string, common.SWARMSIZE)

	// split originalData into slicedData
	for i := 0; i < k; i++ {
		slicedData[i] = string(originalData[i*bytesPerSlice : i*bytesPerSlice+bytesPerSlice])
	}

	// split redundantString into slicedData
	for i := k; i < common.SWARMSIZE; i++ {
		slicedData[i] = redundantString[(i-k)*bytesPerSlice : ((i-k)*bytesPerSlice)+bytesPerSlice]
	}

	// free the memory allocated by the C file
	C.free(unsafe.Pointer(redundantChunk))

	return
}

func RebuildBlock(untaintedSlices []string, sliceIndicies []uint8, k int, bytesPerSlice int) (originalData []byte, err error) {
	// check for legal size of k and m
	m := common.SWARMSIZE - k
	if k > common.SWARMSIZE || k < 1 {
		err = fmt.Errorf("k must be greater than 0 but smaller than %v", common.SWARMSIZE)
		return
	}

	// check for legal size of bytesPerSlice
	if bytesPerSlice < common.MINSLICESIZE || bytesPerSlice > common.MAXSLICESIZE {
		err = fmt.Errorf("bytesPerSlice must be greater than %v and smaller than %v", common.MINSLICESIZE, common.MAXSLICESIZE)
		return
	}

	// check that input data is correct number of slices
	if len(untaintedSlices) != k {
		err = fmt.Errorf("there must be k elements in untaintedSlices")
		return
	}

	// check that input indicies are correct number of indicies
	if len(sliceIndicies) != k {
		err = fmt.Errorf("there must be k elements in sliceIndicies")
		return
	}

	// move all data into a single slice for C
	originalData = make([]byte, 0, k*bytesPerSlice)
	for slice := range untaintedSlices {
		byteSlice := []byte(untaintedSlices[slice])

		// verify that each string is the correct length
		if len(byteSlice) != bytesPerSlice {
			err = fmt.Errorf("at least 1 of 'untaintedSlices' is the wrong length")
			return
		}

		originalData = append(originalData, byteSlice...)
	}

	// call the recovery option
	C.recoverData(C.int(k), C.int(m), C.int(bytesPerSlice), (*C.uchar)(unsafe.Pointer(&originalData[0])), (*C.uchar)(unsafe.Pointer(&sliceIndicies[0])))

	return
}
