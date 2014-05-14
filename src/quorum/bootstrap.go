package quorum

import (
	"common"
	"common/crypto"
	"fmt"
)

// Announce ourself to the bootstrap address, who will announce us to the quorum
func (s *State) JoinSia() (err error) {
	m := &common.Message{
		Dest: bootstrapAddress,
		Proc: "State.HandleJoinSia",
		Args: *s.self,
		Resp: nil,
	}
	s.messageRouter.SendAsyncMessage(m)
	return
}

// Adds a new Participants, and then announces them with their index
// Currently not safe - Participants need to be added during compile()
func (s *State) HandleJoinSia(p Participant, arb *struct{}) (err error) {
	// find index for Participant
	s.participantsLock.Lock()
	i := 0
	for i = 0; i < common.QuorumSize; i++ {
		if s.participants[i] == nil {
			break
		}
	}
	s.participantsLock.Unlock()
	p.index = byte(i)
	err = s.AddNewParticipant(p, nil)
	if err != nil {
		return
	}

	// see if the quorum is full
	if i == common.QuorumSize {
		return fmt.Errorf("failed to add Participant")
	}

	// now announce a new Participant at index i
	s.broadcast(&common.Message{
		Proc: "State.AddNewParticipant",
		Args: p,
		Resp: nil,
	})
	return
}

// Add a Participant to the state, tell the Participant about ourselves
func (s *State) AddNewParticipant(p Participant, arb *struct{}) (err error) {
	if int(p.index) > len(s.participants) {
		err = fmt.Errorf("Corrupt Input")
		return
	}

	s.participantsLock.RLock()
	if s.participants[p.index] != nil {
		s.participantsLock.RUnlock()
		return
	}
	s.participantsLock.RUnlock()
	// for this Participant, make the heartbeat map and add the default heartbeat
	hb := new(heartbeat)
	s.heartbeatsLock.Lock()
	s.participantsLock.Lock()
	s.heartbeats[p.index] = make(map[crypto.TruncatedHash]*heartbeat)
	s.heartbeats[p.index][emptyHash] = hb
	s.heartbeatsLock.Unlock()

	compare := p.compare(s.self)
	if compare == true {
		// add our self object to the correct index in Participants
		s.self.index = p.index
		s.participants[p.index] = s.self
		s.tickingLock.Lock()
		s.ticking = true
		s.tickingLock.Unlock()
		go s.tick()
	} else {
		// add the Participant to Participants
		s.participants[p.index] = &p

		// tell the new guy about ourselves
		s.messageRouter.SendAsyncMessage(&common.Message{
			Dest: p.address,
			Proc: "State.AddNewParticipant",
			Args: *s.self,
			Resp: nil,
		})
	}
	s.participantsLock.Unlock()
	return
}
