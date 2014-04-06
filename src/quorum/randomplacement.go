package quorum

import (
	"common"
)

func RandomPlacement(toPlace int, SEED string) []int {
	ascii := []byte(SEED)
	intAscii := make([]int, len(ascii))
	buckets := make([]int, common.QUORUMSIZE)
	tmp := toPlace
	index := 0
	ascInd := 0
	//splits toPlace into the buckets according to given SEED
	for tmp != 0 {
		index += intAscii[ascInd]
		if index >= common.QUORUMSIZE {
			index = index - common.QUORUMSIZE
		}
		buckets[index] += 1
		tmp--
		ascInd++
		if ascInd == len(intAscii) {
			ascInd = 0
		}
	}
	return buckets
}
