package swarm

import (
	"common"
)

func RandomPlacement(toPlace int, SEED string) []int {
	ascii := []byte(SEED)
	intAscii := make([]int, len(ascii))

	buckets := make([]int, common.REQUIREDTOFILL)

	tmp := toPlace
	index := 0
	ascInd := 0
	//splits toPlace into the buckets (REQUIREDTOFILL) according
	//to the seed given
	for tmp != 0 {
		index += intAscii[ascInd]
		if index >= common.REQUIREDTOFILL {
			index = index - common.REQUIREDTOFILL
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
