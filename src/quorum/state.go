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
	joinSia byte = iota
	incomingSignedHeartbeat
	addressChangeNotification
	newParticipant
)

// Bootstrapping
const (
	bootstrapId   common.Identifier = 0
	bootstrapHost string            = "localhost"
	bootstrapPort int               = 9988
)

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
	self             *participant                    // ourselves
	participantIndex byte                            // our participant index
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

// participant to string
func (p *participant) marshal() (mp []byte) {
	ma := p.address.Marshal()
	mp = append(ma, p.publicKey[:]...)
	return
}

// string to participant
func unmarshalParticipant(mp []byte) (p *participant, err error) {
	if len(mp) < crypto.PublicKeySize+5 {
		err = fmt.Errorf("Length of mp is too small to be a participant")
		return
	}

	p = new(participant)
	copy(p.publicKey[:], mp[len(mp)-crypto.PublicKeySize:])
	addy, err := common.UnmarshalAddress(mp[:len(mp)-crypto.PublicKeySize])
	p.address = *addy
	return
}

// Create and initialize a state object.
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
	s.self.address = messageRouter.AddMessageHandler(s)
	s.self.publicKey = pubKey
	s.secretKey = secKey
	for i := range s.previousEntropyStage1 {
		s.previousEntropyStage1[i] = emptyHash
	}
	s.participantIndex = 255
	s.currentStep = 1
	s.wallets = make(map[string]uint64)

	return
}

// Announce ourself to the bootstrap address, who will announce us to the quorum
func (s *State) JoinSia() (err error) {
	// Send join message to bootstrap address
	m := new(common.Message)
	m.Destination.Id = 0
	m.Destination.Host = "localhost"
	m.Destination.Port = 9988
	m.Payload = append([]byte(string(joinSia)), s.self.marshal()...)
	err = s.messageRouter.SendMessage(m)
	return
}

// Called by the MessageRouter in case of an address change
func (s *State) SetAddress(addr *common.Address) {
	s.participantsLock.Lock()
	s.participants[s.participantIndex].address = *addr
	s.participantsLock.Unlock()

	// now notifiy everyone else in the quorum that the address has changed:
	// that will consist of a 'moved locations' message that has been signed
}

// receives a message and determines what function will handle it.
func (s *State) HandleMessage(m []byte) {
	// message type is stored in the first byte, switch on this type
	switch m[0] {
	case incomingSignedHeartbeat:
		s.handleSignedHeartbeat(m[1:])
	case joinSia:
		s.handleJoinSia(m[1:])
	case addressChangeNotification:
		s.updateParticipant(m[1:])
	case newParticipant:
		s.addNewParticipant(m[1:])
	default:
		log.Infoln("Got message of unrecognized type")
	}
}

// Adds a new participants, and then announces them with their index
func (s *State) handleJoinSia(payload []byte) {
	// find index for participant
	s.participantsLock.Lock()
	i := 0
	for i = 0; i < common.QuorumSize; i++ {
		if s.participants[i] == nil {
			var err error
			s.participants[i], err = unmarshalParticipant(payload)
			if err != nil {
				// log perhaps?... still need to figure out error handling in Sia
				return
			}
			break
		}
	}
	s.participantsLock.Unlock()

	// see if the quorum is full
	if i == common.QuorumSize {
		return
	}

	// now announce a new participant at index i
	var header [2]byte
	header[0] = byte(newParticipant)
	header[1] = byte(i)
	payload = append(header[:], payload...)
	s.broadcast(payload)
}

// A participant can update their address, etc. at any time
func (s *State) updateParticipant(msp []byte) {
	// this message is actually a signature of a participant
	// it's valid if the signature matches the public key
}

// Add a participant to the state, tell the participant about ourselves
func (s *State) addNewParticipant(payload []byte) {
	// extract index and participant object from payload
	participantIndex := payload[0]
	p, err := unmarshalParticipant(payload[1:])
	if err != nil {
		return
	}

	// for this participant, make the heartbeat map and add the default heartbeat
	hb := new(heartbeat)
	emptyHash, err := crypto.CalculateTruncatedHash(hb.entropyStage2[:])
	hb.entropyStage1 = emptyHash
	s.heartbeatsLock.Lock()
	s.participantsLock.Lock()
	s.heartbeats[participantIndex] = make(map[crypto.TruncatedHash]*heartbeat)
	s.heartbeats[participantIndex][emptyHash] = hb
	s.heartbeatsLock.Unlock()

	if *p == *s.self {
		// add our self object to the correct index in participants
		s.participants[participantIndex] = s.self
		s.participantIndex = participantIndex
		s.tickingLock.Lock()
		s.ticking = true
		s.tickingLock.Unlock()

		go s.tick()
	} else {
		// add the participant to participants
		s.participants[participantIndex] = p

		// tell the new guy about ourselves
		m := new(common.Message)
		m.Destination = p.address
		m.Payload = append([]byte(string(newParticipant)), s.self.marshal()...)
		s.messageRouter.SendMessage(m)
	}
	s.participantsLock.Unlock()
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
//
// needs to be converted to return uint64
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
