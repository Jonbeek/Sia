#include "longhair/include/cauchy_256.h"
#include <stdlib.h>
#include <stdio.h>
#include <string.h>

// encodeRedundancy takes as input a 'k', the number of nonredundant segments
// and an 'm', the number of redundant segments. k + m must be less than 256.
// bytesPerSegment is the size of each segment, and then originalBlock is a pointer
// to the original data, which is assumed to be of size k * bytesPerSegment
//
// The return value is a block of data m * bytesPerSegment which contains all of
// the redundant data. The data does not get segmented into pieces in this
// function.
//
// encodeRedundancy does not do any error checking, all of that must happen
// in the calling function.
static char *encodeRedundancy(int k, int m, int bytesPerSegment, char *originalBlock) {
	// verify that correct library is linked
	if(cauchy_256_init()) {
		fprintf(stderr, "Failed to init cauchy_256 library");
		exit(1);
	}

	// break original data into segments
	const unsigned char *originalSegments[k];
	int i;
	for(i = 0; i < k; i++) {
		originalSegments[i] = &originalBlock[i * bytesPerSegment];
	}

	// allocate space for redundant segments
	unsigned char *redundantSegments = calloc(sizeof(unsigned char), m * bytesPerSegment);

	// encode the redundant segments
	if(cauchy_256_encode(k, m, originalSegments, redundantSegments, bytesPerSegment)) {
		fprintf(stderr, "Failed to encode pieces - strange...");
		exit(1);
	}

	return redundantSegments;
}

// recoverData takes as input 'k', the number of nonredundant segments and 'm',
// the number of redundant segments. 'bytesPerSegment' indicates how large each
// segment is. remainingSegments is a pointer to a block of data that contains
// exactly 'k' uncorrupted segments. 'remainingSegmentIndicies' indicate which
// segments of the original set the uncorrupted ones correspond with.
//
// The data is edited and sorted in place. Upon returning, 'remainingSegments'
// will be the original data in order.
static void recoverData(int k, int m, int bytesPerSegment, unsigned char *remainingSegments, unsigned char *remainingSegmentIndicies) {
	// verify that correct library is linked
	if(cauchy_256_init()) {
		//error
		exit(1);
	}

	// create blocks around data using the indicies
	Block segmentInfo[k];
	int i, j;
	for(i = 0; i < k; i++) {
		segmentInfo[i].data = &remainingSegments[i * bytesPerSegment];
		segmentInfo[i].row = remainingSegmentIndicies[i];
	}
	
	// decode redundant segments into original segments
	if(cauchy_256_decode(k, m, segmentInfo, bytesPerSegment)) {
		//error
		exit(1);
	}

	/* sort memory back into original order */
	/* because I want to push sooner rather than later, I'm using an n^2 sort */
	/* eventually, I'll implement something more efficient */

	// allocate space to copy memory into
	char tempData[bytesPerSegment];
	char tempIndex;

	// insertion sort (kinda)
	i = 0;
	while(i < k) {
		while(segmentInfo[i].row == i && i < k) {
			i++;
		}

		if(i == k) {
			break;
		}

		// copy [i] into temp
		tempIndex = segmentInfo[i].row;
		memcpy(tempData, segmentInfo[i].data, bytesPerSegment);

		j = i;
		while(segmentInfo[j].row != i) {
			j++;
		}

		// copy [j] into [i]
		segmentInfo[i].row = segmentInfo[j].row;
		memcpy(segmentInfo[i].data, segmentInfo[j].data, bytesPerSegment);

		// copy temp into [j]
		segmentInfo[j].row = tempIndex;
		memcpy(segmentInfo[j].data, tempData, bytesPerSegment);
	}
}
