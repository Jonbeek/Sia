package log

import (
	"time"
	"runtime"
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

func itoa(i, width uint) []byte {
	buf := make([]byte)
	if i == 0 && width == 0 {
		buf = append(buf, '0')
		return buf
	}
	var b [32]byte
	bp := len(b)
	for ; i > 0 || width > 0; i /= 10 {
		width--
		bp--
		b[bp] = byte(i % 10) + '0'
	}
	buf = append(buf, b[bp:])
	return buf
}

func (r record) Format() []byte {
	buf := make([]byte)
	// Copied from the official log.
	// Date writing
	year, month, day := r.logtime.Date()
	buf = append(buf, itoa(year, 4)...)
	buf = append(buf, '/')
	buf = append(buf, itoa(month, 2)...)
	buf = append(buf, '/')
	buf = append(buf, itoa(day, 2)...)
	buf = append(buf, ' ')
	// Time writing
	hour, min, sec := r.logtime.Clock()
	buf = append(buf, itoa(hour, 2)...)
	buf = append(buf, ':')
	buf = append(buf, itoa(min, 2)...)
	buf = append(buf, ':')
	buf = append(buf, itoa(sec, 2)...)
	buf = append(buf, '.')
	nanosec := r.logtime.Nanosecond()/1e3
	buf = append(buf, itoa(nanosec, 6))
	buf = append(buf, ' ')
	// Priority writing
	switch r.priority {
	case Perror:
		buf = append(buf, "[ERROR] "...)
	case Pprint:
		// ...no idea
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
	buf = append(buf, r.file)
	buf = append(buf, ':')
	buf = append(buf, itoa(r.line, 0)
	buf = append(buf, ": "...)
	// Write the message!
	buf = append(buf, r.message)
	// Add a newline if not included
	// Somewhat copied from official log
	if len(buf) > 0 && buf[len(buf)-1] != '\n' {
		buf = append(buf, '\n')
	}
	return buf
}


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
