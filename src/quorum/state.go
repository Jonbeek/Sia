package quorum

import (
	"common"
	"common/crypto"
	"common/log"
	"fmt"
	"sync"
)

// Message Types
const (
	joinQuorumRequest uint8 = iota
	incomingSignedHeartbeat
)

// Leaves space for flexibility in the future
type participantIndex uint8

// Identifies other members of the quorum
type participant struct {
	address   common.Address
	publicKey crypto.PublicKey
}

// The state provides persistence to the consensus algorithms. Every participant
// should have an identical state.
type State struct {
	// Network Variables
	messageRouter    common.MessageRouter
	participants     [common.QuorumSize]*participant // list of participants
	participantsLock sync.RWMutex                    // write-locks for compile only
	participantIndex participantIndex                // our participant index
	secretKey        crypto.SecretKey

	// Heartbeat Variables
	storedEntropyStage2 common.Entropy // hashed to EntropyStage1 for previous heartbeat

	// Compile Variables
	previousEntropyStage1 [common.QuorumSize]crypto.TruncatedHash // used to verify the next round of heartbeats
	currentEntropy        common.Entropy                          // Used to generate random numbers during compilation
	upcomingEntropy       common.Entropy                          // Used to compute entropy for next block

	// Consensus Algorithm Status
	currentStep    int
	stepLock       sync.RWMutex // prevents a benign race condition
	ticking        bool
	tickingLock    sync.Mutex
	heartbeats     [common.QuorumSize]map[crypto.TruncatedHash]*heartbeat
	heartbeatsLock sync.Mutex

	// Wallet Data
	wallets map[string]uint64
}

// Create and initialize a state object. Crypto keys are not created until a quorum is joined
func CreateState(messageRouter common.MessageRouter) (s State, err error) {
	// check that we have a non-nil messageSender
	if messageRouter == nil {
		err = fmt.Errorf("Cannot initialize with a nil messageRouter")
		return
	}

	// calculate the value of an empty hash (default for storedEntropyStage2 on all hosts is a blank array)
	emptyHash, err := crypto.CalculateTruncatedHash(s.storedEntropyStage2[:])
	if err != nil {
		return
	}

	// set state variables to their defaults
	s.messageRouter = messageRouter
	for i := range s.previousEntropyStage1 {
		s.previousEntropyStage1[i] = emptyHash
	}
	s.participantIndex = 255
	s.currentStep = 1
	s.wallets = make(map[string]uint64)

	return
}

// Take an unstarted State and begin the consensus algorithm cycle
func (s *State) Start() (err error) {
	// state cannot call Start() if it has already started
	s.tickingLock.Lock()
	defer s.tickingLock.Unlock()

	// if s.ticking == true, then Start() was called but _ (end()?) was not
	if s.ticking {
		fmt.Errorf("State is ticking, cannot Start()")
		return
	}

	// create first heartbeat
	hb, err := s.newHeartbeat()
	if err != nil {
		return
	}
	heartbeatHash, err := crypto.CalculateTruncatedHash([]byte(hb.marshal()))
	if err != nil {
		return
	}

	// add heartbeat to our map
	s.heartbeatsLock.Lock()
	s.heartbeats[s.participantIndex][heartbeatHash] = hb
	s.heartbeatsLock.Unlock()

	// sign and broadcast heartbeat
	sh, err := s.signHeartbeat(hb)
	if err != nil {
		return
	}
	s.announceSignedHeartbeat(sh)

	// start ticking
	s.ticking = true
	go s.tick()
	return
}

// Called by the MessageRouter in case of an address change
func (s *State) SetAddress(addr *common.Address) {
	s.participantsLock.Lock()
	s.participants[s.participantIndex].address = *addr
	s.participantsLock.Unlock()

	// now notifiy everyone else in the quorum that the address has changed:
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

// Takes a payload and sends it in a message to every participant in the quorum
func (s *State) broadcast(payload []byte) {
	s.participantsLock.RLock()
	for i := range s.participants {
		if s.participants[i] != nil {
			m := new(common.Message)
			m.Payload = payload
			m.Destination = s.participants[i].address
			err := s.messageRouter.SendMessage(m)
			if err != nil {
				log.Errorln("messageSender returning an error")
			}
		}
	}
	s.participantsLock.RUnlock()
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
