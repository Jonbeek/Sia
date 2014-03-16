package common

import (
	"container/heap"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
)

const (
	Perror = 1 << iota
	Pprint
	Pwarning
	Pinfo
	Pdebug
	PstdFlags = Perror | Pprint | Pwarning
)

type record struct {
	logtime  time.Time
	priority uint
	message  string
	// Records the source file and line of the calling function
	file string
	line int
}

func newRecord(calldepth int, priority uint, message string) *record {
	// Calldepth in the Go log is 2, but this isn't the Go log.
	// For the PriorityLog, it's 3.
	// priority should be ONE AND ONLY ONE of the constants above.
	r := new(record)
	// Record as early as possible for most accurate timing
	r.logtime = time.Now()
	r.priority = priority
	r.message = message
	var ok bool
	// According to the comments in the Go log, this takes a while.
	_, r.file, r.line, ok = runtime.Caller(calldepth)
	if !ok {
		r.file = "???"
		r.line = 0
	}
	return r
}

func (r record) Less(q record) bool {
	return r.logtime.Before(q.logtime)
}

func (r record) Format() []byte {
	// TODO: Implement this
	return nil
}

// Does not need to be concurrent safe: PriorityLog handles concurrency
type recordHeap struct {
	recs []*record
}

func (rh recordHeap) Len() int {
	return len(rh.recs)
}

func (rh recordHeap) Less(i, j int) bool {
	return rh.recs[i].Less(*rh.recs[j])
}

func (rh *recordHeap) Swap(i, j int) {
	rh.recs[i], rh.recs[j] = rh.recs[j], rh.recs[i]
}

func (rh *recordHeap) Clear() {
	rh.recs = rh.recs[:]
}

func (rh *recordHeap) Push(x interface{}) {
	rh.recs = append(rh.recs, x.(*record))
}
func (rh *recordHeap) Pop() interface{} {
	n := len(rh.recs)
	x := rh.recs[n-1]
	rh.recs = rh.recs[:n-1]
	return x
}

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
	// Lock is used to prevent one goroutine from continuously making entries
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

func (pl *PriorityLog) Error(v ..interface{}) {
	pl.Claim()
	defer pl.Unclaim()
	go pl.log(Perror, fmt.Sprint(v...))
}

func (pl *PriorityLog) Print(v ...interface{}) {
	pl.Claim()
	defer pl.Unclaim()
	go pl.log(Pprint, fmt.Sprint(v...))
}

func (pl *PriorityLog) Warning(v ..interface{}) {
	pl.Claim()
	defer pl.Unclaim()
	go pl.log(Pwarning, fmt.Sprint(v...))
}

func (pl *PriorityLog) Info(v ..interface{}) {
	pl.Claim()
	defer pl.Unclaim()
	go pl.log(Pinfo, fmt.Sprint(v...))
}

func (pl *PriorityLog) Debug(v ..interface{}) {
	pl.Claim()
	defer pl.Unclaim()
	go pl.log(Pdebug, fmt.Sprint(v...))
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

var std = NewPriorityLog(os.Stderr, PstdFlags, true)

func LogStored() {
	std.LogStored()
}

func Clear() {
	std.Clear()
}

func Error(v ..interface{}) {
	std.Claim()
	defer std.Unclaim()
	go std.log(Perror, fmt.Sprint(v...))
}

func Print(v ...interface{}) {
	std.Claim()
	defer std.Unclaim()
	go std.log(Pprint, fmt.Sprint(v...))
}

func Warning(v ..interface{}) {
	std.Claim()
	defer std.Unclaim()
	go std.log(Pwarning, fmt.Sprint(v...))
}

func Info(v ..interface{}) {
	std.Claim()
	defer std.Unclaim()
	go std.log(Pinfo, fmt.Sprint(v...))
}

func Debug(v ..interface{}) {
	std.Claim()
	defer std.Unclaim()
	go std.log(Pdebug, fmt.Sprint(v...))
}
