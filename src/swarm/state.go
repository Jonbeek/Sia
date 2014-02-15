package swarm

type State struct {
	// swarm information - id, network location, etc.
	// connection information to all hosts in swarm
	// connection information to all hosts in parent swarm

	// wallet database

	// block history, snapshot history (if there is a snapshot history)

	// scheduled scripts (once we write those in)

	// active heartbeat - being updated by transactions and such

	DRNGSeed []byte

	// The data used to produce Stage1 hashes in the recent heartbeat
	SecretEntropy   []byte
	SecretFileProof []byte

	ActiveBlock Block
}
