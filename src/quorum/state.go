package quorum

import (
	"bytes"
	"common"
	"common/crypto"
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
	ID:   1,
	Host: "localhost",
	Port: 9988,
}

// empty hash value
var emptyEntropy = common.Entropy{}
var emptyHash, _ = crypto.CalculateTruncatedHash(emptyEntropy[:])

// Identifies other members of the quorum
type Participant struct {
	index     byte
	address   common.Address
	publicKey *crypto.PublicKey
}

// The state provides persistence to the consensus algorithms. Every participant
// should have an identical state.
type State struct {
	// Network Variables
	messageRouter    common.MessageRouter
	participants     [common.QuorumSize]*Participant // list of participants
	participantsLock sync.RWMutex                    // write-locks for compile only
	self             *Participant                    // ourselves
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
}

// Returns true if the values of the participants are equivalent
func (p0 *Participant) compare(p1 *Participant) bool {
	// false if either participant is nil
	if p0 == nil || p1 == nil {
		return false
	}

	// return false if the addresses are not equal
	if p0.address != p1.address {
		return false
	}

	// return false if the public keys are not equivalent
	compare := p0.publicKey.Compare(p1.publicKey)
	if compare != true {
		return false
	}

	return true
}

func (p *Participant) GobEncode() (gobParticipant []byte, err error) {
	// Error checking for nil values
	if p == nil {
		err = fmt.Errorf("Cannot encode nil value p")
		return
	}
	if p.publicKey == nil {
		err = fmt.Errorf("Cannot encode nil value p.publicKey")
		return
	}
	epk := (*ecdsa.PublicKey)(p.publicKey)
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

func (p *Participant) GobDecode(gobParticipant []byte) (err error) {
	if p == nil {
		err = fmt.Errorf("Cannot decode into nil Participant")
		return
	}

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

	// initialize State with default values and keypair
	s = &State{
		messageRouter: messageRouter,
		self: &Participant{
			index:     255,
			address:   messageRouter.Address(),
			publicKey: pubKey,
		},
		secretKey:   secKey,
		currentStep: 1,
	}

	// register State and store our assigned ID
	s.self.address.ID = messageRouter.RegisterHandler(s)

	// initialize entropy stage1 to the emptyHash
	for i := range s.previousEntropyStage1 {
		s.previousEntropyStage1[i] = emptyHash
	}

	// a call to joinSia() may be placed here... behavior not fully defined
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

// Adds a new Participants, and then announces them with their index
// Currently not safe - Participants need to be added during compile()
func (s *State) HandleJoinSia(p Participant, arb *struct{}) (err error) {
	// find index for Participant
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
		return fmt.Errorf("failed to add Participant")
	}

	p.index = byte(i)
	// now announce a new Participant at index i
	s.broadcast(&common.Message{
		Proc: "State.AddNewParticipant",
		Args: p,
		Resp: nil,
	})
	return
}

// A Participant can update their address, etc. at any time
func (s *State) updateParticipant(msp []byte) {
	// this message is actually a signature of a Participant
	// it's valid if the signature matches the public key
}

// Add a Participant to the state, tell the Participant about ourselves
func (s *State) AddNewParticipant(p Participant, arb *struct{}) (err error) {
	// for this Participant, make the heartbeat map and add the default heartbeat
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
			Args: s.self,
			Resp: nil,
		})
	}
	s.participantsLock.Unlock()
	return
}

// Takes a Message and broadcasts it to every Participant in the quorum
func (s *State) broadcast(m *common.Message) {
	s.participantsLock.RLock()
	for i := range s.participants {
		if s.participants[i] != nil {
			nm := *m
			nm.Dest = s.participants[i].address
			s.messageRouter.SendAsyncMessage(&nm)
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
