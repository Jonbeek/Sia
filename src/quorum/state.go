package quorum

import (
	"common"
	"common/crypto"
	"common/log"
	"fmt"
	"sync"
)

type participantIndex int

// The state provides persistence to the consensus algorithms. Every participant
// should have an identical state.
type State struct {
	lock sync.Mutex

	// Network Variables
	messageSender common.MessageSender
	participants  [common.QuorumSize]*Participant // list of participants

	// Our information
	secretKey        crypto.SecretKey // public key in our participant index
	participantIndex participantIndex // our participant index

	// Heartbeat Variables
	storedEntropyStage2 common.Entropy // hashed to EntropyStage1 for previous heartbeat

	// Compile Variables
	previousEntropyStage1 [common.QuorumSize]crypto.TruncatedHash // used to verify the next round of heartbeats
	currentEntropy        common.Entropy                          // Used to generate random numbers during compilation
	upcomingEntropy       common.Entropy                          // Used to compute entropy for next block

	// Consensus Algorithm Status
	currentStep int
	ticking     bool
	tickLock    sync.Mutex
	heartbeats  [common.QuorumSize]map[crypto.TruncatedHash]*heartbeat

	// Wallet Data
	wallets map[string]uint64
}

type Participant struct {
	Address   common.Address
	PublicKey crypto.PublicKey
}

// Create and initialize a state object
func CreateState(messageSender common.MessageSender, participantIndex participantIndex) (s State, err error) {
	// check that participantIndex is legal, then add basic info
	if int(participantIndex) >= common.QuorumSize {
		err = fmt.Errorf("Invalid participant index!")
		return
	}
	s.participantIndex = participantIndex
	s.messageSender = messageSender

	// initialize crypto keys
	pubKey, secKey, err := crypto.CreateKeyPair()
	if err != nil {
		return
	}
	s.secretKey = secKey

	// create and fill out participant object, add it to our list of participants
	self := new(Participant)
	self.Address = messageSender.Address()
	self.Address.Id = common.Identifier(participantIndex)
	self.PublicKey = pubKey
	s.AddParticipant(self, participantIndex)

	// intialize remaining values to their defaults
	s.currentStep = 1
	s.wallets = make(map[string]uint64)
	emptyHash, err := crypto.CalculateTruncatedHash(s.storedEntropyStage2[:])
	if err != nil {
		return
	}
	for i := range s.previousEntropyStage1 {
		s.previousEntropyStage1[i] = emptyHash
	}

	return
}

// self() fetches the state's participant object
func (s *State) Self() (p *Participant) {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.participants[s.participantIndex]
}

// add participant to s.Participants, and initialize the heartbeat map
func (s *State) AddParticipant(p *Participant, i participantIndex) (err error) {
	s.lock.Lock()
	// Check that there is not already a participant for the index
	if s.participants[i] != nil {
		err = fmt.Errorf("A participant already exists for the given index!")
		return
	}
	s.participants[i] = p

	// initialize the heartbeat map for this participant
	s.heartbeats[i] = make(map[crypto.TruncatedHash]*heartbeat)
	s.lock.Unlock()

	return
}

// Use the entropy stored in the state to generate a random integer [low, high)
func (s *State) randInt(low int, high int) (randInt int, err error) {
	// verify there's a gap between the numbers
	if low == high {
		err = fmt.Errorf("low and high cannot be the same number")
		return
	}

	// Convert CurrentEntropy into an int
	rollingInt := 0
	for i := 0; i < 4; i++ {
		rollingInt = rollingInt << 4
		rollingInt += int(s.currentEntropy[0])
	}

	randInt = (rollingInt % (high - low)) + low

	// Convert random number seed to next value
	truncatedHash, err := crypto.CalculateTruncatedHash(s.currentEntropy[:])
	s.currentEntropy = common.Entropy(truncatedHash)
	return
}

func (s *State) HandleMessage(m []byte) {
	// message type is stored in the first byte, switch on this type
	switch m[0] {
	case 1:
		s.lock.Lock()
		s.handleSignedHeartbeat(m[1:])
		s.lock.Unlock()
	default:
		log.Infoln("Got message of unrecognized type")
	}
}

func (s *State) Identifier() common.Identifier {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.participants[s.participantIndex].Address.Id
}

// Take an unstarted State and begin the consensus algorithm cycle
func (s *State) Start() {
	// start the ticker to progress the state
	go s.tick()

	s.lock.Lock()
	// create first heartbeat and add it to heartbeat map, then announce it
	hb, err := s.newHeartbeat()
	if err != nil {
		return
	}
	heartbeatHash, err := crypto.CalculateTruncatedHash([]byte(hb.marshal()))
	s.heartbeats[s.participantIndex][heartbeatHash] = hb
	shb, err := s.signHeartbeat(hb)
	if err != nil {
		return
	}
	s.announceSignedHeartbeat(shb)
	s.lock.Unlock()
}
