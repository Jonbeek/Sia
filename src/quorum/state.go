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
	// a temporary overall lock, will eventually be replaced with component locks
	lock sync.Mutex

	// Network Variables
	messageSender    common.MessageSender
	participants     [common.QuorumSize]*participant // list of participants
	participantIndex participantIndex                // our participant index
	secretKey        crypto.SecretKey                // public key in our participant index

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

// Only temporarily a public object, will eventually be 'type participant struct'
// makes building easier since we don't have a 'join swarm' function yet
type participant struct {
	address   common.Address
	publicKey crypto.PublicKey
}

// Create and initialize a state object
func CreateState(messageSender common.MessageSender, participantIndex participantIndex) (s State, err error) {
	// check that we have a non-nil messageSender
	if messageSender == nil {
		err = fmt.Errorf("Cannot initialize with a nil messageSender")
		return
	}

	// check that participantIndex is legal
	if int(participantIndex) >= common.QuorumSize {
		err = fmt.Errorf("Invalid participant index!")
		return
	}

	// initialize crypto keys
	pubKey, secKey, err := crypto.CreateKeyPair()
	if err != nil {
		return
	}

	// create and fill out the participant object
	self := new(participant)
	self.address = messageSender.Address()
	self.address.Id = common.Identifier(participantIndex)
	self.publicKey = pubKey

	// calculate the value of an empty hash (default for storedEntropyStage2 on all hosts is a blank array)
	emptyHash, err := crypto.CalculateTruncatedHash(s.storedEntropyStage2[:])
	if err != nil {
		return
	}

	// set state variables to their defaults
	s.messageSender = messageSender
	s.AddParticipant(self, participantIndex)
	s.secretKey = secKey
	for i := range s.previousEntropyStage1 {
		s.previousEntropyStage1[i] = emptyHash
	}
	s.participantIndex = participantIndex
	s.currentStep = 1
	s.wallets = make(map[string]uint64)

	return
}

// self() fetches the state's participant object
func (s *State) Self() (p *participant) {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.participants[s.participantIndex]
}

// add participant to s.Participants, and initialize the heartbeat map
func (s *State) AddParticipant(p *participant, i participantIndex) (err error) {
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
	return s.participants[s.participantIndex].address.Id
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
