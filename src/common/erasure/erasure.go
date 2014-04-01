package erasure

// #cgo LDFLAGS: longhair/bin/liblonghair.a
// #include "bridge.c"
import "C"

import (
	"common"
	"fmt"
	"unsafe"
)

func EncodeRing(originalSlices []byte, m int, sliceSize int) (redundantSlices []string, err error) {
	// check that 'm' is legal
	k := common.SWARMSIZE - m
	if k <= 0 || k >= common.SWARMSIZE {
		err = fmt.Errorf("m must be greater than 0 and smaller than %v", common.SWARMSIZE)
		return
	}

	// check that sliceSize is not to big or small
	if sliceSize < common.MINSLICESIZE || sliceSize > common.MAXSLICESIZE {
		err = fmt.Errorf("sliceSize must be greater than %v and smaller than %v", common.MINSLICESIZE, common.MAXSLICESIZE)
		return
	}

	// check that sliceSize is divisible by 8
	if sliceSize%8 != 0 {
		err = fmt.Errorf("sliceSize must be divisible by 8")
		return
	}

	// check that originalSlices is the correct size
	if len(originalSlices) != sliceSize*k {
		err = fmt.Errorf("originalSlices incorrectly padded, must be of size 'sliceSize' * %v - 'm'", common.SWARMSIZE)
		return
	}

	// call c library to encode data
	redundantChunk := C.encodeRedundancy(C.int(k), C.int(m), C.int(sliceSize), (*C.char)(unsafe.Pointer(&originalSlices[0])))
	redundantString := C.GoStringN(redundantChunk, C.int(m*sliceSize))

	// redundantChunk into redundantSlices
	redundantSlices = make([]string, m)
	for i := 0; i < m; i++ {
		redundantSlices[i] = redundantString[i*sliceSize : (i*sliceSize)+sliceSize]
	}

	// free the memory allocated by the C file
	// I'm not sure at this point where everything is pointing,
	// this might also kill the redundant slices
	C.free(unsafe.Pointer(redundantChunk))

	return
}
