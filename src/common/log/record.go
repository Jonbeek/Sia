package log

import (
	"runtime"
	"time"
)

type record struct {
	logtime  time.Time
	priority uint
	message  string
	// Records the source file and line of the calling function
	file string
	line int
}

func newRecord(calldepth int, now time.Time, priority uint, message string) *record {
	// Calldepth in the Go log is 2, but this isn't the Go log.
	// For the PriorityLog, it's 3.
	// priority should be ONE AND ONLY ONE of the constants above.
	r := new(record)
	// Record as early as possible for most accurate timing
	r.logtime = now
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
	buf := make([]byte, 40)
	// Date writing
	year, month, day := r.logtime.Date()
	itoa(&buf, year, 4)
	buf = append(buf, '/')
	// Month is its own type.
	itoa(&buf, int(month), 2)
	buf = append(buf, '/')
	itoa(&buf, day, 2)
	buf = append(buf, ' ')
	// Time writing
	hour, min, sec := r.logtime.Clock()
	itoa(&buf, hour, 2)
	buf = append(buf, ':')
	itoa(&buf, min, 2)
	buf = append(buf, ':')
	itoa(&buf, sec, 2)
	buf = append(buf, '.')
	nanosec := r.logtime.Nanosecond() / 1e3
	itoa(&buf, nanosec, 6)
	buf = append(buf, ' ')
	// Priority writing
	switch r.priority {
	case Pfatal:
		buf = append(buf, "[FATAL] "...)
	case Perror:
		buf = append(buf, "[ERROR] "...)
	case Pwarning:
		buf = append(buf, "[WARNING] "...)
	case Pinfo:
		buf = append(buf, "[INFO] "...)
	case Pdebug:
		buf = append(buf, "[DEBUG] "...)
	default:
		// Assume RAM not ECC, not really a major error
		buf = append(buf, "[UNKNOWN] "...)
	}
	// Write the whole filepath because no options (yet)
	buf = append(buf, r.file...)
	buf = append(buf, ':')
	itoa(&buf, r.line, 0)
	buf = append(buf, ": "...)
	// Write the message!
	buf = append(buf, r.message...)
	// Add a newline if not included
	if len(buf) > 0 && buf[len(buf)-1] != '\n' {
		buf = append(buf, '\n')
	}
	return buf
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
