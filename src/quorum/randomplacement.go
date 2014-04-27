package quorum

import (
	"common"
)

func (s *State) RandomPlacement(toPlace int) (buckets []int, err error) {
	buckets = make([]int, common.QuorumSize)
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
