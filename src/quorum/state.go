package quorum

import (
	"common"
	"common/crypto"
	"fmt"
	"sync"
)

type ParticipantIndex uint8

// The state provides persistence to the consensus algorithms. Every participant
// should have an identical state.
type State struct {
	// Network Variables
	MessageSender common.MessageSender
	Participants  [common.QuorumSize]*Participant // list of participants

	// Our information
	SecretKey        crypto.SecretKey // public key in our participant index
	ParticipantIndex ParticipantIndex // our participant index

	// Heartbeat Variables
	StoredEntropyStage2 common.Entropy // hashed to EntropyStage1 for previous heartbeat

	// Compile Variables
	PreviousEntropyStage1 [common.QuorumSize]crypto.TruncatedHash // used to verify the next round of heartbeats
	CurrentEntropy        common.Entropy                          // Used to generate random numbers during compilation
	UpcomingEntropy       common.Entropy                          // Used to compute entropy for next block

	// Consensus Algorithm Status
	CurrentStep int
	Ticking     bool
	TickLock    sync.Mutex
	Heartbeats  [common.QuorumSize]map[crypto.TruncatedHash]*Heartbeat

	// Wallet Data
	Wallets map[string]uint64
}

type Participant struct {
	Address   common.Address
	PublicKey crypto.PublicKey
}

// Create and initialize a state object
func CreateState(messageSender common.MessageSender, participantIndex ParticipantIndex) (s State, err error) {
	// check that participantIndex is legal, then add basic info
	if int(participantIndex) >= common.QuorumSize {
		err = fmt.Errorf("Invalid participant index!")
		return
	}
	s.ParticipantIndex = participantIndex
	s.MessageSender = messageSender

	// initialize crypto keys
	pubKey, secKey, err := crypto.CreateKeyPair()
	if err != nil {
		return
	}
	s.SecretKey = secKey

	// create and fill out participant object, add it to our list of participants
	self := new(Participant)
	self.Address = messageSender.Address()
	self.Address.Id = common.Identifier(participantIndex)
	self.PublicKey = pubKey
	s.AddParticipant(self, participantIndex)

	// intialize remaining values to their defaults
	s.CurrentStep = 1
	s.Wallets = make(map[string]uint64)
	emptyHash, err := crypto.CalculateTruncatedHash(s.StoredEntropyStage2[:])
	if err != nil {
		return
	}
	for i := range s.PreviousEntropyStage1 {
		s.PreviousEntropyStage1[i] = emptyHash
	}

	return
}

// self() fetches the state's participant object
func (s *State) Self() (p *Participant) {
	return s.Participants[s.ParticipantIndex]
}

// add participant to s.Participants, and initialize the heartbeat map
func (s *State) AddParticipant(p *Participant, i ParticipantIndex) (err error) {
	// Check that there is not already a participant for the index
	if s.Participants[i] != nil {
		err = fmt.Errorf("A participant already exists for the given index!")
		return
	}
	s.Participants[i] = p

	// initialize the heartbeat map for this participant
	s.Heartbeats[i] = make(map[crypto.TruncatedHash]*Heartbeat)

	return
}

// Use the entropy stored in the state to generate a random integer [low, high)
func (s *State) RandInt(low int, high int) (randInt int, err error) {
	// verify there's a gap between the numbers
	if low == high {
		err = fmt.Errorf("low and high cannot be the same number")
		return
	}

	// Convert CurrentEntropy into an int
	rollingInt := 0
	for i := 0; i < 4; i++ {
		rollingInt = rollingInt << 4
		rollingInt += int(s.CurrentEntropy[0])
	}

	randInt = (rollingInt % (high - low)) + low

	// Convert random number seed to next value
	truncatedHash, err := crypto.CalculateTruncatedHash(s.CurrentEntropy[:])
	s.CurrentEntropy = common.Entropy(truncatedHash)
	return
}

func (s *State) HandleMessage(m []byte) {
	// message type is stored in the first byte, switch on this type
	println(s.ParticipantIndex, ": got a message: ", m[0])
	switch m[0] {
	case 1:
		s.HandleSignedHeartbeat(m[1:])
	default:
		println("got dud message")
	}
}

func (s *State) Identifier() common.Identifier {
	return s.Participants[s.ParticipantIndex].Address.Id
}

// Take an unstarted State and begin the consensus algorithm cycle
func (s *State) Start() {
	// start the ticker to progress the state
	go s.Tick()

	// create first heartbeat and add it to heartbeat map, then announce it
	hb, err := s.NewHeartbeat()
	if err != nil {
		return
	}
	heartbeatHash, err := crypto.CalculateTruncatedHash([]byte(hb.Marshal()))
	s.Heartbeats[s.ParticipantIndex][heartbeatHash] = hb
	shb, err := s.SignHeartbeat(hb)
	if err != nil {
		return
	}
	s.announceSignedHeartbeat(shb)
}
