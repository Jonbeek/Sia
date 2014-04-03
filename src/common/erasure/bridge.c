#include "longhair/include/cauchy_256.h"
#include <stdlib.h>
#include <stdio.h>
#include <string.h>

static char *encodeRedundancy(int k, int m, int bytesPerSlice, char *originalBlock) {
	// verify that correct library is linked
	if(cauchy_256_init()) {
		fprintf(stderr, "Failed to init cauchy_256 library");
		exit(1);
	}

	// break original data into slices
	const unsigned char *originalSlices[k];
	int i;
	for(i = 0; i < k; i++) {
		originalSlices[i] = &originalBlock[i * bytesPerSlice];
	}

	// allocate space for redundant slices
	unsigned char *redundantSlices = calloc(sizeof(unsigned char), m * bytesPerSlice);

	// encode the redundant slices
	if(cauchy_256_encode(k, m, originalSlices, redundantSlices, bytesPerSlice)) {
		fprintf(stderr, "Failed to encode pieces - strange...");
		exit(1);
	}

	return redundantSlices;
}

static void recoverData(int k, int m, int bytesPerSlice, unsigned char *remainingSlices, unsigned char *remainingSliceIndicies) {
	// verify that correct library is linked
	if(cauchy_256_init()) {
		//error
		exit(1);
	}

	// create blocks around data using the indicies
	Block sliceInfo[k];
	int i, j;
	for(i = 0; i < k; i++) {
		sliceInfo[i].data = &remainingSlices[i * bytesPerSlice];
		sliceInfo[i].row = remainingSliceIndicies[i];
	}
	
	// decode redundant slices into original slices
	if(cauchy_256_decode(k, m, sliceInfo, bytesPerSlice)) {
		//error
		exit(1);
	}

	/* sort memory back into original order */
	/* because I want to push sooner rather than later, I'm using an n^2 sort */
	/* eventually, I'll probably implement a radix sort */

	// allocate space to copy memory into
	char tempData[bytesPerSlice];
	char tempIndex;

	// insertion sort (kinda)
	i = 0;
	while(i < k) {
		while(sliceInfo[i].row == i && i < k) {
			i++;
		}

		if(i == k) {
			break;
		}

		// copy [i] into temp
		tempIndex = sliceInfo[i].row;
		memcpy(tempData, sliceInfo[i].data, bytesPerSlice);

		j = i;
		while(sliceInfo[j].row != i) {
			j++;
		}

		// copy [j] into [i]
		sliceInfo[i].row = sliceInfo[j].row;
		memcpy(sliceInfo[i].data, sliceInfo[j].data, bytesPerSlice);

		// copy temp into [j]
		sliceInfo[j].row = tempIndex;
		memcpy(sliceInfo[j].data, tempData, bytesPerSlice);
	}
}
