package quorum

import (
	"common"
	"fmt"
)

func (s *State) RandomPlacement(toPlace int) (buckets []int, err error) {
	buckets = make([]int, common.QuorumSize)
	if toPlace < 0 {
		return buckets, fmt.Errorf("Cannot place a negative number!")
	}
	for toPlace != 0 {
		rand, err := s.randInt(0, common.QuorumSize)
		if err != nil {
			return buckets, err
		}
		buckets[rand]++
		toPlace--
	}
	return buckets, nil
}
