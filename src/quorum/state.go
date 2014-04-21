package quorum

import (
	"common"
	"common/crypto"
	"common/log"
	"fmt"
	"sync"
)

const (
	joinQuorumRequest       = 2
	incomingSignedHeartbeat = 1
)

type participantIndex int

// The state provides persistence to the consensus algorithms. Every participant
// should have an identical state.
type State struct {
	// Network Variables
	messageSender    common.MessageSender
	participants     [common.QuorumSize]*participant // list of participants
	participantsLock sync.RWMutex                    // write-locks for compile only
	participantIndex participantIndex                // our participant index
	secretKey        crypto.SecretKey                // public key in our participant index

	// Heartbeat Variables
	storedEntropyStage2 common.Entropy // hashed to EntropyStage1 for previous heartbeat

	// Compile Variables
	previousEntropyStage1 [common.QuorumSize]crypto.TruncatedHash // used to verify the next round of heartbeats
	currentEntropy        common.Entropy                          // Used to generate random numbers during compilation
	upcomingEntropy       common.Entropy                          // Used to compute entropy for next block

	// Consensus Algorithm Status
	currentStep    int
	ticking        bool
	tickLock       sync.Mutex
	heartbeats     [common.QuorumSize]map[crypto.TruncatedHash]*heartbeat
	heartbeatsLock sync.Mutex

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
func CreateState(messageSender common.MessageSender) (s State, err error) {
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

	// calculate the value of an empty hash (default for storedEntropyStage2 on all hosts is a blank array)
	emptyHash, err := crypto.CalculateTruncatedHash(s.storedEntropyStage2[:])
	if err != nil {
		return
	}

	// set state variables to their defaults
	s.messageSender = messageSender
	s.secretKey = secKey
	for i := range s.previousEntropyStage1 {
		s.previousEntropyStage1[i] = emptyHash
	}
	s.participantIndex = 255
	s.currentStep = 1
	s.wallets = make(map[string]uint64)

	return
}

// receives a message and determines what function will handle it.
// HandleMessage is not responsible for mutexes
func (s *State) HandleMessage(m []byte) {
	// message type is stored in the first byte, switch on this type
	switch m[0] {
	case incomingSignedHeartbeat:
		s.handleSignedHeartbeat(m[1:])
	case joinQuorumRequest:
		// the message is going to contain connection information
		// will need to return a marshalled state
	default:
		log.Infoln("Got message of unrecognized type")
	}
}

// self() fetches the state's participant object
func (s *State) Self() (p participant) {
	// check that we have joined a quorum, otherwise we have no participant object
	if participantIndex == 255 {
		return
	}

	s.participantsLock.RLock()
	p = s.participants[s.participantIndex]
	s.participantsLock.RUnlock()
	return
}

// Use the entropy stored in the state to generate a random integer [low, high)
// randInt only runs during compile(), when the mutexes are already locked
func (s *State) randInt(low int, high int) (randInt int, err error) {
	// verify there's a gap between the numbers
	if low == high {
		err = fmt.Errorf("low and high cannot be the same number")
		return
	}

	// Convert CurrentEntropy into an int
	rollingInt := 0
	for i := 0; i < 4; i++ {
		rollingInt = rollingInt << 8
		rollingInt += int(s.currentEntropy[i])
	}

	randInt = (rollingInt % (high - low)) + low

	// Convert random number seed to next value
	truncatedHash, err := crypto.CalculateTruncatedHash(s.currentEntropy[:])
	s.currentEntropy = common.Entropy(truncatedHash)
	return
}

// The network server many need to request the identifier
func (s *State) Identifier() (i common.Identifier) {
	s.participantsLock.RLock()
	i = s.participants[s.participantIndex].address.Id
	s.participantsLock.RUnlock()
	return
}

// Take an unstarted State and begin the consensus algorithm cycle
func (s *State) Start() {
	// state cannot call Start() if it has already started
	s.tickLock.Lock()
	if s.ticking {
		s.tickLock.Unlock()
		return
	} else {
		s.ticking = true
	}
	s.tickLock.Unlock()

	// start the ticker to progress the state
	go s.tick()

	// create first heartbeat and add it to heartbeat map, then announce it
	hb, err := s.newHeartbeat()
	if err != nil {
		return
	}
	heartbeatHash, err := crypto.CalculateTruncatedHash([]byte(hb.marshal()))

	s.heartbeatsLock.Lock()
	s.heartbeats[s.participantIndex][heartbeatHash] = hb
	s.heartbeatsLock.Unlock()

	shb, err := s.signHeartbeat(hb)
	if err != nil {
		return
	}
	s.announceSignedHeartbeat(shb)
}
