package log

import (
	"container/heap"
	"fmt"
	"io"
	"sync"
)

const (
	Perror = 1 << iota
	Pprint
	Pwarning
	Pinfo
	Pdebug
	PstdFlags = Perror | Pprint | Pwarning
)

// PriorityLog is a log with priority levels.
// Upon receiving a request to log something, the PriorityLog will:
// Check to see if the priority level is set in now.
// If so, print it immediately.
// Else, check to see if dispose is set
// If it is, ignore the request.
// Else, store it in the recordHeap
type PriorityLog struct {
	// now control the levels of priority which are immediately printed
	// and stored to print in case of an error.
	dispose bool
	now     uint
	records *recordHeap
	out     io.Writer
	// lock protects the above fields
	lock sync.Mutex
	// stop stops creation of new log entries.
	// THere are two mutexes solely to prevent deadlocks.
	stop sync.Mutex
	// Every logging function counts as two in pending
	// The first for formatting the message, the second for creating the record
	pending sync.WaitGroup
}

func NewPriorityLog(out io.Writer, flags uint, dispose bool) *PriorityLog {
	pl := new(PriorityLog)
	pl.out = out
	pl.now = flags
	pl.dispose = dispose
	pl.records = new(recordHeap)
	return pl
}

func (pl *PriorityLog) Claim() {
	// Must be done before calls to log
	// Lock is used to prevent one routine from continuously making entries
	pl.stop.Lock()
	defer pl.stop.Unlock()
	// One for the caller, one for log
	pl.pending.Add(2)
}

func (pl *PriorityLog) Unclaim() {
	// No need to wait, finishing less harmful than starting
	pl.pending.Done()
}

func (pl *PriorityLog) log(priority uint, message string) {
	// Calling function called Claim, hopefully
	defer pl.Unclaim()
	// The calldepth will always be 3, if called directly
	rec := newRecord(3, priority, message)
	// Claim complete use over the variables of pl
	pl.lock.Lock()
	defer pl.lock.Unlock()
	if pl.now&priority != 0 {
		pl.out.Write(rec.Format())
	} else {
		if !pl.dispose {
			heap.Push(pl.records, rec)
		}
		// If dispose is set, then it's a waste of effort
		// But not memory, at least not long lasting memory.
	}
}

// SetGlobalFlags sets the the default flag for which priority levels are
// defered and which ones are handled immediately
func (pl *PriorityLog) SetGlobalFlags(newflags uint) {
	pl.lock.Lock()
	defer pl.lock.Unlock()
	pl.now = newflags
}

// SetDeferedBehavior changes how defered log entries are handled.
// (i.e. disposed of or not)
func (pl *PriorityLog) SetDeferedBehavior(dispose bool) {
	pl.lock.Lock()
	defer pl.lock.Unlock()
	pl.dispose = dispose
}

func (pl *PriorityLog) Error(v ...interface{}) {
	pl.Claim()
	defer pl.Unclaim()
	pl.log(Perror, fmt.Sprint(v...))
}

func (pl *PriorityLog) Print(v ...interface{}) {
	pl.Claim()
	defer pl.Unclaim()
	pl.log(Pprint, fmt.Sprint(v...))
}

func (pl *PriorityLog) Warning(v ...interface{}) {
	pl.Claim()
	defer pl.Unclaim()
	pl.log(Pwarning, fmt.Sprint(v...))
}

func (pl *PriorityLog) Info(v ...interface{}) {
	pl.Claim()
	defer pl.Unclaim()
	pl.log(Pinfo, fmt.Sprint(v...))
}

func (pl *PriorityLog) Debug(v ...interface{}) {
	pl.Claim()
	defer pl.Unclaim()
	pl.log(Pdebug, fmt.Sprint(v...))
}

func (pl *PriorityLog) LogStored() {
	// Prints and clears the log
	// Stop making new log entries
	pl.stop.Lock()
	defer pl.stop.Unlock()
	// Wait for any yet to be added log entries
	pl.pending.Wait()
	pl.lock.Lock()
	defer pl.lock.Unlock()
	// Write it all out
	for pl.records.Len() > 0 {
		rec := heap.Pop(pl.records).(*record)
		pl.out.Write(rec.Format())
	}
}

func (pl *PriorityLog) Clear() {
	// Uncertain if this should wait for all pending records or not.
	// For now, don't bother.
	pl.lock.Lock()
	defer pl.lock.Unlock()
	pl.records.Clear()
}
