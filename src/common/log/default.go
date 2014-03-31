package log

import (
	"fmt"
	"os"
	"time"
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
	std.log(time.Now(), Pfatal, fmt.Sprint(v...))
	os.Exit(1)
}

func Error(v ...interface{}) {
	std.Claim()
	defer std.Unclaim()
	std.log(time.Now(), Perror, fmt.Sprint(v...))
}

func Warning(v ...interface{}) {
	std.Claim()
	defer std.Unclaim()
	std.log(time.Now(), Pwarning, fmt.Sprint(v...))
}

func Info(v ...interface{}) {
	std.Claim()
	defer std.Unclaim()
	std.log(time.Now(), Pinfo, fmt.Sprint(v...))
}

func Debug(v ...interface{}) {
	std.Claim()
	defer std.Unclaim()
	std.log(time.Now(), Pdebug, fmt.Sprint(v...))
}

func Fatalln(v ...interface{}) {
	std.Claim()
	defer std.Unclaim()
	std.log(time.Now(), Pfatal, fmt.Sprintln(v...))
	os.Exit(1)
}

func Errorln(v ...interface{}) {
	std.Claim()
	defer std.Unclaim()
	std.log(time.Now(), Perror, fmt.Sprintln(v...))
}

func Warningln(v ...interface{}) {
	std.Claim()
	defer std.Unclaim()
	std.log(time.Now(), Pwarning, fmt.Sprintln(v...))
}

func Infoln(v ...interface{}) {
	std.Claim()
	defer std.Unclaim()
	std.log(time.Now(), Pinfo, fmt.Sprintln(v...))
}

func Debugln(v ...interface{}) {
	std.Claim()
	defer std.Unclaim()
	std.log(time.Now(), Pdebug, fmt.Sprintln(v...))
}

func Fatalf(format string, v ...interface{}) {
	std.Claim()
	defer std.Unclaim()
	std.log(time.Now(), Pfatal, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func Errorf(format string, v ...interface{}) {
	std.Claim()
	defer std.Unclaim()
	std.log(time.Now(), Perror, fmt.Sprintf(format, v...))
}

func Warningf(format string, v ...interface{}) {
	std.Claim()
	defer std.Unclaim()
	std.log(time.Now(), Pwarning, fmt.Sprintf(format, v...))
}

func Infof(format string, v ...interface{}) {
	std.Claim()
	defer std.Unclaim()
	std.log(time.Now(), Pinfo, fmt.Sprintf(format, v...))
}

func Debugf(format string, v ...interface{}) {
	std.Claim()
	defer std.Unclaim()
	std.log(time.Now(), Pdebug, fmt.Sprintf(format, v...))
}
