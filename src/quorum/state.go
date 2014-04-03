package quorum

import (
	"common"
)

type State struct {
}

func (s *State) HandleUpdate(u common.Update) {
}

// Input: some structure indicating which state we are joining
// Output: a state object
//
// Or does the state just enter it's own mainloop() sort of deal?
// I really think that network should be sitting on a stack of states,
// and then it just calls 'go state.Handle()' and then you have a mutex
// to make sure that Handle() isn't causing any memory problems.
func JoinState() (s *State) {
	return
}
