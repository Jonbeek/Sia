package log

import (
	"fmt"
	"os"
)

var std = NewPriorityLog(os.Stderr, PstdFlags, true)

func LogStored() {
	std.LogStored()
}

func Clear() {
	std.Clear()
}

// SetGlobalFlags calls SetGlobalFlags on the standard logger
func SetGlobalFlags(newflags uint) {
	std.SetGlobalFlags(newflags)
}

// SetDeferedBehavior calls SetDeferedBehavior on the standard logger
func SetDeferedBehavior(dispose bool) {
	std.SetDeferedBehavior(dispose)
}

func Fatal(v ...interface{}) {
	std.Claim()
	defer std.Unclaim()
	std.log(Perror, fmt.Sprint(v...))
	os.Exit(1)
}

func Error(v ...interface{}) {
	std.Claim()
	defer std.Unclaim()
	std.log(Perror, fmt.Sprint(v...))
}

func Print(v ...interface{}) {
	std.Claim()
	defer std.Unclaim()
	std.log(Pprint, fmt.Sprint(v...))
}

func Warning(v ...interface{}) {
	std.Claim()
	defer std.Unclaim()
	std.log(Pwarning, fmt.Sprint(v...))
}

func Info(v ...interface{}) {
	std.Claim()
	defer std.Unclaim()
	std.log(Pinfo, fmt.Sprint(v...))
}

func Debug(v ...interface{}) {
	std.Claim()
	defer std.Unclaim()
	std.log(Pdebug, fmt.Sprint(v...))
}

func Println(v ...interface{}) {
	std.Claim()
	defer std.Unclaim()
	std.log(Pprint, fmt.Sprintln(v...))
}
