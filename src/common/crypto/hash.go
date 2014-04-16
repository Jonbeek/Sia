package crypto

// #include <sodium.h>
import "C"

import (
	"fmt"
	"unsafe"
)

// Uses the hash in libsodium
func CalculateHash(data []byte) (hash Hash, err error) {
	hashPointer := (*C.uchar)(unsafe.Pointer(&hash[0]))
	messagePointer := (*C.uchar)(unsafe.Pointer(&data[0]))
	sizeOfMessage := C.ulonglong(len(data))
	success := C.crypto_hash(hashPointer, messagePointer, sizeOfMessage)
	if success != 0 {
		err = fmt.Errorf("Error in calculating hash")
	}

	return
}

// Calls Hash, and then returns only the first TruncatedHashSize bytes
func CalculateTruncatedHash(data []byte) (tHash TruncatedHash, err error) {
	hash, err := CalculateHash(data)
	if err != nil {
		return
	}

	copy(tHash[:], hash[:])
	return
}
