package swarm

import (
	"common"
)

type StateSteady struct {
}

func (s *StateSteady) HandleUpdate(u common.Update) State {
	return s
}

func NewStateSteady() State {
	return &StateSteady{}
}
