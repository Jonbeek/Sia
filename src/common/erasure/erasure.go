package erasure

// #cgo LDFLAGS: /home/david/git/Sia/src/common/erasure/longhair/bin/liblonghair.a -lstdc++
// #include "bridge.c"
import "C"

import (
	"common"
	"fmt"
	"unsafe"
)

func EncodeRing(originalData []byte, m int, bytesPerSlice int) (redundantSlices []string, err error) {
	// check that 'm' is legal
	k := common.SWARMSIZE - m
	if k <= 0 || k >= common.SWARMSIZE {
		err = fmt.Errorf("m must be greater than 0 and smaller than %v", common.SWARMSIZE)
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
	redundantChunk := C.encodeRedundancy(C.int(k), C.int(m), C.int(bytesPerSlice), (*C.char)(unsafe.Pointer(&originalData[0])))
	redundantString := C.GoStringN(redundantChunk, C.int(m*bytesPerSlice))

	// redundantChunk into redundantSlices
	redundantSlices = make([]string, m)
	for i := 0; i < m; i++ {
		redundantSlices[i] = redundantString[i*bytesPerSlice : (i*bytesPerSlice)+bytesPerSlice]
	}

	// free the memory allocated by the C file
	C.free(unsafe.Pointer(redundantChunk))

	return
}

func RebuildRing(untaintedSlices [][]byte, k int, bytesPerSlice int) (originalData []byte) {
	return
}
