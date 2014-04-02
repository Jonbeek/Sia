#include "longhair/include/cauchy_256.h"
#include <stdlib.h>

static char *encodeRedundancy(int k, int m, int bytesPerSlice, char *originalBlock) {
	if(cauchy_256_init()) {
		// error
		exit(1);
	}

	const unsigned char *originalSlices[k];

	int i;
	for(i = 0; i < k; i++) {
		originalSlices[i] = &originalBlock[i * bytesPerSlice];
	}

	unsigned char *redundantSlices = calloc(sizeof(unsigned char), m * bytesPerSlice);

	if(cauchy_256_encode(k, m, originalSlices, redundantSlices, bytesPerSlice)) {
		// error
		exit(1);
	}

	return redundantSlices;
}

// edits remainingSlices in place.
static void recoverData(int k, int m, int bytesPerSlice, char *remainingSlices, int *remainingSliceIndicies) {
	if(cauchy_256_init()) {
		//error
		exit(1);
	}

	Block sliceInfo[k];

	// build the blocks from remainingSlices and remainingSliceIndicies
	int i;
	for(i = 0; i < k; i++) {
		sliceInfo[i].data = &remainingSlices[i*bytesPerSlice];
		sliceInfo[i].row = remainingSliceIndicies[i];
	}
	
	if(cauchy_256_decode(k, m, sliceInfo, bytesPerSlice)) {
		//error
		exit(1);
	}
}
