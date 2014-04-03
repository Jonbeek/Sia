#include "longhair/include/cauchy_256.h"
#include <stdlib.h>
#include <stdio.h>
#include <string.h>

// encodeRedundancy takes as input a 'k', the number of nonredundant slices
// and an 'm', the number of redundant slices. k + m must be less than 256.
// bytesPerSlice is the size of each slice, and then originalBlock is a pointer
// to the original data, which is assumed to be of size k * bytesPerSlice
//
// The return value is a block of data m * bytesPerSlice which contains all of
// the redundant data. The data does not get segmented into pieces in this
// function.
//
// encodeRedundancy does not do any error checking, all of that must happen
// in the calling function.
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

// recoverData takes as input 'k', the number of nonredundant slices and 'm',
// the number of redundant slices. 'bytesPerSlice' indicates how large each
// slice is. remainingSlices is a pointer to a block of data that contains
// exactly 'k' uncorrupted slices. 'remainingSliceIndicies' indicate which
// slices of the original set the uncorrupted ones correspond with.
//
// The data is edited and sorted in place. Upon returning, 'remainingSlices'
// will be the original data in order.
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
	/* eventually, I'll implement something more efficient */

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
