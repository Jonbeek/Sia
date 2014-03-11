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
	// Every log helper function counts as two in pending
	// The first for formatting, the second for creating the record
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

func (pl *PriorityLog) log(priority uint, message string) {
	// The calldepth will always be 3, if called directly.
	pl.pending.Add(1)
	defer pl.pending.Done()
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
	}
}

func (pl *PriorityLog) Print(v ...interface{}) {
	pl.pending.Add(1)
	defer pl.pending.Done()
	go pl.log(Pprint, fmt.Sprint(v...))
}

func (pl *PriorityLog) LogStored() {
	// Prints and clears the log
	pl.pending.Wait()
	pl.lock.Lock()
	defer pl.lock.Unlock()
	for pl.records.Len() > 0 {
		rec := heap.Pop(pl.records).(*record)
		pl.out.Write(rec.Format())
	}
}

func (pl *PriorityLog) Clear() {
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

func Print(v ...interface{}) {
	std.pending.Add(1)
	defer std.pending.Done()
	go std.log(Pprint, fmt.Sprint(v...))
}
