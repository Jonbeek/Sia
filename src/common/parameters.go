package common

import (
	"time"
)

// swarmsize is the number of hosts managing each set of files
var SWARMSIZE int = 192

var STATEINFORMEDDELTA time.Duration = 1 * time.Second

const (
	// How many bytes of entropy must be produced each entropy cycle
	ENTROPYVOLUME int = 32

	// How big a single slice of data is for a host, in bytes
	MINSLICESIZE int = 512
	MAXSLICESIZE int = 1048576 // 1 MB
)
