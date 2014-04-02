#include "longhair/include/cauchy_256.h"
#include <stdlib.h>

static char *encodeRedundancy(int k, int m, int sliceSize, char *originalBlock) {
	if(cauchy_256_init()) {
		// error
		exit(1);
	}

	const unsigned char *originalSlices[k];

	int i;
	for(i = 0; i < k; i++) {
		originalSlices[i] = &originalBlock[i * sliceSize];
	}

	unsigned char *redundantSlices = calloc(sizeof(unsigned char), m * sliceSize);

	if(cauchy_256_encode(k, m, originalSlices, redundantSlices, sliceSize)) {
		exit(1);
	}

	return redundantSlices;
}

static char *recoverData(int k, int m, int sliceSize, char *remainingSlices) {
	if(cauchy_256_init()) {
		//error
		exit(1);
	}
}
