package quorum

import (
	"bytes"
	"common"
	"common/crypto"
	"common/log"
	"crypto/ecdsa"
	"encoding/gob"
	"fmt"
	"sync"
)

// Message Types
const (
	joinSia byte = iota
	incomingSignedHeartbeat
	addressChangeNotification
	newParticipant
)

// Bootstrapping
var bootstrapAddress = common.Address{
	ID:   0,
	Host: "localhost",
	Port: 9988,
}

// Identifies other members of the quorum
type participant struct {
	index     byte
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
	self             *participant                    // ourselves
	secretKey        crypto.SecretKey                // our secret key

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

// Returns true if the values of the participants are equivalent
func (p0 *participant) compare(p1 *participant) bool {
	// false if either participant is nil
	if p0 == nil || p1 == nil {
		return false
	}

	// return false if the addresses are not equal
	if p0.address != p1.address {
		return false
	}

	// return false if the public keys are not equivalent
	compare := p0.publicKey.Compare(&p1.publicKey)
	if compare != true {
		return false
	}

	return true
}

func (p *participant) GobEncode() (gobParticipant []byte, err error) {
	// Error checking for nil values
	epk := ecdsa.PublicKey(p.publicKey)
	if epk.X == nil {
		err = fmt.Errorf("Cannot encode nil value p.publicKey.X")
		return
	}
	if epk.Y == nil {
		err = fmt.Errorf("Cannot encode nil value p.publicKey.Y")
		return
	}

	// Encoding the participant
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err = encoder.Encode(p.address)
	if err != nil {
		return
	}
	err = encoder.Encode(p.publicKey)
	if err != nil {
		return
	}
	gobParticipant = w.Bytes()
	return
}

func (p *participant) GobDecode(gobParticipant []byte) (err error) {
	r := bytes.NewBuffer(gobParticipant)
	decoder := gob.NewDecoder(r)
	err = decoder.Decode(&p.address)
	if err != nil {
		return
	}
	err = decoder.Decode(&p.publicKey)
	if err != nil {
		return
	}
	return
}

// Create and initialize a state object. Set everything to default.
func CreateState(messageRouter common.MessageRouter) (s *State, err error) {
	s = new(State)
	// check that we have a non-nil messageSender
	if messageRouter == nil {
		err = fmt.Errorf("Cannot initialize with a nil messageRouter")
		return
	}

	// create a signature keypair for this state
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
	s.messageRouter = messageRouter
	s.self = new(participant)
	s.self.index = 255
	s.self.address = messageRouter.Address()
	s.self.address.ID = messageRouter.RegisterHandler(s)
	s.self.publicKey = pubKey
	s.secretKey = secKey
	for i := range s.previousEntropyStage1 {
		s.previousEntropyStage1[i] = emptyHash
	}
	s.currentStep = 1
	s.wallets = make(map[string]uint64)

	return
}

// Announce ourself to the bootstrap address, who will announce us to the quorum
func (s *State) JoinSia() (err error) {
	m := &common.Message{
		Dest: bootstrapAddress,
		Proc: "State.HandleJoinSia",
		Args: *s.self,
		Resp: nil,
	}
	err = s.messageRouter.SendMessage(m)
	return
}

// Called by the MessageRouter in case of an address change
func (s *State) SetAddress(addr *common.Address) {
	s.participantsLock.Lock()
	s.participants[s.self.index].address = *addr
	s.participantsLock.Unlock()

	// now notifiy everyone else in the quorum that the address has changed:
	// that will consist of a 'moved locations' message that has been signed
}

// Adds a new participants, and then announces them with their index
// Currently not safe - participants need to be added during compile()
func (s *State) HandleJoinSia(p participant, arb *struct{}) (err error) {
	// find index for participant
	s.participantsLock.Lock()
	i := 0
	for i = 0; i < common.QuorumSize; i++ {
		if s.participants[i] == nil {
			s.participants[i] = &p
			break
		}
	}
	s.participantsLock.Unlock()

	// see if the quorum is full
	if i == common.QuorumSize {
		return fmt.Errorf("failed to add participant")
	}

	p.index = byte(i)
	// now announce a new participant at index i
	s.broadcast(&common.Message{
		Proc: "State.AddNewParticipant",
		Args: p,
		Resp: nil,
	})
	return
}

// A participant can update their address, etc. at any time
func (s *State) updateParticipant(msp []byte) {
	// this message is actually a signature of a participant
	// it's valid if the signature matches the public key
}

// Add a participant to the state, tell the participant about ourselves
func (s *State) AddNewParticipant(p participant, arb *struct{}) (err error) {
	// for this participant, make the heartbeat map and add the default heartbeat
	hb := new(heartbeat)
	emptyHash, err := crypto.CalculateTruncatedHash(hb.entropyStage2[:])
	if err != nil {
		return
	}
	hb.entropyStage1 = emptyHash
	s.heartbeatsLock.Lock()
	s.participantsLock.Lock()
	s.heartbeats[p.index] = make(map[crypto.TruncatedHash]*heartbeat)
	s.heartbeats[p.index][emptyHash] = hb
	s.heartbeatsLock.Unlock()

	compare := p.compare(s.self)
	if compare == true {
		// add our self object to the correct index in participants
		s.self.index = p.index
		s.participants[p.index] = s.self
		s.tickingLock.Lock()
		s.ticking = true
		s.tickingLock.Unlock()
		go s.tick()
	} else {
		// add the participant to participants
		s.participants[p.index] = &p

		// tell the new guy about ourselves
		s.messageRouter.SendMessage(&common.Message{
			Dest: p.address,
			Proc: "State.AddNewParticipant",
			Args: s.self,
			Resp: nil,
		})
	}
	s.participantsLock.Unlock()
	return
}

// Takes a Message and broadcasts it to every participant in the quorum
func (s *State) broadcast(m *common.Message) {
	s.participantsLock.RLock()
	for i := range s.participants {
		if s.participants[i] != nil {
			m.Dest = s.participants[i].address
			err := s.messageRouter.SendMessage(m)
			if err != nil {
				log.Infoln(err)
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
