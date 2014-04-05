package log

import (
	"container/heap"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

const (
	Pfatal = 1 << iota
	Perror
	Pwarning
	Pinfo
	Pdebug
	PstdFlags = Pfatal | Perror | Pwarning
	// Use only in the main program.
	PdebugFlags = Pfatal | Perror | Pwarning | Pinfo | Pdebug
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

func (pl *PriorityLog) log(now time.Time, priority uint, message string) {
	// Calling function called Claim, hopefully
	defer pl.Unclaim()
	// Claim complete use over the variables of pl
	pl.lock.Lock()
	defer pl.lock.Unlock()
	// Optimization!
	if pl.now&priority != 0 || !pl.dispose {
		pl.lock.Unlock()
		// Takes a while.
		// The calldepth will always be 3, if called directly
		rec := newRecord(3, now, priority, message)
		pl.lock.Lock()
		if pl.now&priority != 0 {
			pl.out.Write(rec.Format())
		} else {
			heap.Push(pl.records, rec)
		}
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

func (pl *PriorityLog) Fatal(v ...interface{}) {
	pl.Claim()
	defer pl.Unclaim()
	pl.log(time.Now(), Pfatal, fmt.Sprint(v...))
	os.Exit(1)
}

func (pl *PriorityLog) Error(v ...interface{}) {
	pl.Claim()
	defer pl.Unclaim()
	pl.log(time.Now(), Perror, fmt.Sprint(v...))
}

func (pl *PriorityLog) Warning(v ...interface{}) {
	pl.Claim()
	defer pl.Unclaim()
	pl.log(time.Now(), Pwarning, fmt.Sprint(v...))
}

func (pl *PriorityLog) Info(v ...interface{}) {
	pl.Claim()
	defer pl.Unclaim()
	pl.log(time.Now(), Pinfo, fmt.Sprint(v...))
}

func (pl *PriorityLog) Debug(v ...interface{}) {
	pl.Claim()
	defer pl.Unclaim()
	pl.log(time.Now(), Pdebug, fmt.Sprint(v...))
}

func (pl *PriorityLog) Fatalln(v ...interface{}) {
	pl.Claim()
	defer pl.Unclaim()
	pl.log(time.Now(), Pfatal, fmt.Sprintln(v...))
	os.Exit(1)
}

func (pl *PriorityLog) Errorln(v ...interface{}) {
	pl.Claim()
	defer pl.Unclaim()
	pl.log(time.Now(), Perror, fmt.Sprintln(v...))
}

func (pl *PriorityLog) Warningln(v ...interface{}) {
	pl.Claim()
	defer pl.Unclaim()
	pl.log(time.Now(), Pwarning, fmt.Sprintln(v...))
}

func (pl *PriorityLog) Infoln(v ...interface{}) {
	pl.Claim()
	defer pl.Unclaim()
	pl.log(time.Now(), Pinfo, fmt.Sprintln(v...))
}

func (pl *PriorityLog) Debugln(v ...interface{}) {
	pl.Claim()
	defer pl.Unclaim()
	pl.log(time.Now(), Pdebug, fmt.Sprintln(v...))
}

func (pl *PriorityLog) Fatalf(format string, v ...interface{}) {
	pl.Claim()
	defer pl.Unclaim()
	pl.log(time.Now(), Pfatal, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func (pl *PriorityLog) Errorf(format string, v ...interface{}) {
	pl.Claim()
	defer pl.Unclaim()
	pl.log(time.Now(), Perror, fmt.Sprintf(format, v...))
}

func (pl *PriorityLog) Warningf(format string, v ...interface{}) {
	pl.Claim()
	defer pl.Unclaim()
	pl.log(time.Now(), Pwarning, fmt.Sprintf(format, v...))
}

func (pl *PriorityLog) Infof(format string, v ...interface{}) {
	pl.Claim()
	defer pl.Unclaim()
	pl.log(time.Now(), Pinfo, fmt.Sprintf(format, v...))
}

func (pl *PriorityLog) Debugf(format string, v ...interface{}) {
	pl.Claim()
	defer pl.Unclaim()
	pl.log(time.Now(), Pdebug, fmt.Sprintf(format, v...))
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
