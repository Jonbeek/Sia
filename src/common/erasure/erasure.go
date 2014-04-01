package erasure

// #cgo CFLAGS:
// #include "bridge.c"
import "C"

import (
	"common"
	"fmt"
	"math"
	//"unsafe"
)

func EncodeRing(input []byte, m int) (output [][]byte, err error) {
	// make sure m is a correct value
	k := common.SWARMSIZE - m
	if k <= 0 || k >= common.SWARMSIZE {
		err = fmt.Errorf("m must be smaller than %v and greater than 0", common.SWARMSIZE)
		return
	}

	// Make sure the input is small enough to fit in one ring
	maxRingSize := common.SWARMSIZE * common.MAXSLICESIZE
	ringSize := len(input) / k * common.SWARMSIZE
	if ringSize > maxRingSize {
		err = fmt.Errorf("input block too large - max size is %v!", maxRingSize)
		return
	}

	// figure out the desired slice size
	// slice size must be divisible by 8
	nonRoundedSliceSize := float64(len(input)) / float64(k)
	nonRoundedSliceSize /= 8
	sliceSize := math.Ceil(nonRoundedSliceSize)
	sliceSize *= 8

	// call c library
	C.encodeRedundancy(C.int(k), C.int(m), C.int(sliceSize), (*C.char)(unsafe.Pointer(&input[0])))
	// need to test that the []byte to char* works correctly

	return
}
